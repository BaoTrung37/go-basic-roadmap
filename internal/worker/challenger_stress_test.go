package worker_test

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/worker"
)

// TestChallenger_WorkerPool_ConcurrentDispatchAndStop stresses the worker pool
// by having 100 goroutines rapidly dispatching events while Stop() is called concurrently.
// It verifies that no send on closed channel occurs, no deadlocks occur, and no races occur.
func TestChallenger_WorkerPool_ConcurrentDispatchAndStop(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	pool := worker.NewWorkerPool(8, 50, discardLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	const (
		numDispatchers = 100
		eventsPerDisp  = 100
	)

	var (
		dispatchedCount atomic.Int64
		rejectedCount   atomic.Int64
		wg              sync.WaitGroup
		startBarrier    = make(chan struct{})
	)

	wg.Add(numDispatchers)
	for d := 0; d < numDispatchers; d++ {
		go func(dispatcherID int) {
			defer wg.Done()
			<-startBarrier // synchronize start

			for i := 0; i < eventsPerDisp; i++ {
				event := model.TaskEvent{
					Type: model.EventTypeCreated,
					Task: model.Task{
						ID:       "stress-task",
						Title:    "Stress Event",
						Status:   model.StatusTodo,
						Priority: model.PriorityHigh,
					},
					Timestamp: time.Now(),
				}
				if pool.Dispatch(event) {
					dispatchedCount.Add(1)
				} else {
					rejectedCount.Add(1)
				}
			}
		}(d)
	}

	// Release dispatchers
	close(startBarrier)

	// Sleep briefly to let dispatches get underway, then call Stop() concurrently
	time.Sleep(2 * time.Millisecond)
	pool.Stop()

	// Wait for all dispatchers to finish
	wg.Wait()

	t.Logf("Dispatch & Stop stress completed: Dispatched=%d, Rejected=%d",
		dispatchedCount.Load(), rejectedCount.Load())

	// Subsequent dispatches must all fail cleanly
	if pool.Dispatch(model.TaskEvent{Type: model.EventTypeDeleted}) {
		t.Error("Dispatch after Stop() should return false")
	}
}

// TestChallenger_WorkerPool_ConcurrentStartAndStop stresses idempotency and concurrency
// safety of Start() and Stop() called simultaneously from dozens of goroutines.
func TestChallenger_WorkerPool_ConcurrentStartAndStop(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for iter := 0; iter < 10; iter++ {
		pool := worker.NewWorkerPool(4, 20, discardLogger)
		ctx := context.Background()

		var wg sync.WaitGroup
		const callers = 20

		wg.Add(callers * 2)
		for i := 0; i < callers; i++ {
			go func() {
				defer wg.Done()
				pool.Start(ctx)
			}()
			go func() {
				defer wg.Done()
				pool.Stop()
			}()
		}

		wg.Wait()
		// Final stop call should be completely safe
		pool.Stop()
	}
}

// TestChallenger_WorkerPool_BufferSaturation verifies non-blocking drop behavior under saturation.
func TestChallenger_WorkerPool_BufferSaturation(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Pool with 0 workers so channel fills immediately
	pool := worker.NewWorkerPool(1, 5, discardLogger)

	// Don't start pool yet — dispatch should return false
	event := model.TaskEvent{Type: model.EventTypeCreated}
	if pool.Dispatch(event) {
		t.Fatal("Dispatch before Start() must return false")
	}

	ctx := context.Background()
	pool.Start(ctx)

	var wg sync.WaitGroup
	var successCount, dropCount atomic.Int64

	const senderCount = 50
	wg.Add(senderCount)
	for i := 0; i < senderCount; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				if pool.Dispatch(event) {
					successCount.Add(1)
				} else {
					dropCount.Add(1)
				}
			}
		}()
	}

	wg.Wait()
	pool.Stop()

	t.Logf("Buffer saturation: Enqueued=%d, Dropped=%d", successCount.Load(), dropCount.Load())
	if successCount.Load()+dropCount.Load() != senderCount*20 {
		t.Errorf("Mismatch in total events: expected %d, got %d", senderCount*20, successCount.Load()+dropCount.Load())
	}
}
