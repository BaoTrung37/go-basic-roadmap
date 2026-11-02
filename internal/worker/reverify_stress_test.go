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

// TestAdversarial_WorkerPool_ConcurrentStartStopDispatchBarrier creates a thundering herd
// where 50 goroutines call Start(), 50 call Stop(), and 100 call Dispatch() simultaneously.
// We repeat this 50 times to catch any non-deterministic race conditions or deadlocks.
func TestAdversarial_WorkerPool_ConcurrentStartStopDispatchBarrier(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for round := 0; round < 50; round++ {
		pool := worker.NewWorkerPool(8, 64, discardLogger)
		ctx, cancel := context.WithCancel(context.Background())

		const (
			starters    = 50
			stoppers    = 50
			dispatchers = 100
		)

		var wg sync.WaitGroup
		wg.Add(starters + stoppers + dispatchers)

		barrier := make(chan struct{})

		// Goroutines calling Start
		for i := 0; i < starters; i++ {
			go func() {
				defer wg.Done()
				<-barrier
				pool.Start(ctx)
			}()
		}

		// Goroutines calling Stop
		for i := 0; i < stoppers; i++ {
			go func() {
				defer wg.Done()
				<-barrier
				pool.Stop()
			}()
		}

		// Goroutines calling Dispatch
		for i := 0; i < dispatchers; i++ {
			go func(id int) {
				defer wg.Done()
				<-barrier
				event := model.TaskEvent{
					Type: model.EventTypeCreated,
					Task: model.Task{
						ID:     "adversarial-task",
						Title:  "Adversarial Stress",
						Status: model.StatusTodo,
					},
					Timestamp: time.Now(),
				}
				_ = pool.Dispatch(event)
			}(i)
		}

		// Unleash all 200 goroutines simultaneously
		close(barrier)

		wg.Wait()
		cancel()
		pool.Stop() // Final cleanup must be idempotent
	}
}

// TestAdversarial_WorkerPool_PreCanceledContext verifies behavior when Start receives an already canceled context.
func TestAdversarial_WorkerPool_PreCanceledContext(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	pool := worker.NewWorkerPool(4, 32, discardLogger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // canceled before Start

	pool.Start(ctx)

	// Dispatch an event
	event := model.TaskEvent{
		Type:      model.EventTypeCreated,
		Task:      model.Task{ID: "pre-canceled"},
		Timestamp: time.Now(),
	}
	_ = pool.Dispatch(event)

	// Stop must cleanly terminate without deadlock
	stopped := make(chan struct{})
	go func() {
		pool.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("Deadlock detected in Stop() after pre-canceled context")
	}
}

// TestAdversarial_WorkerPool_StopBeforeStart verifies that Stop() called before Start()
// correctly marks the pool as stopped and subsequent Start() is a safe no-op.
func TestAdversarial_WorkerPool_StopBeforeStart(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	pool := worker.NewWorkerPool(4, 32, discardLogger)

	// Call Stop first
	pool.Stop()

	// Call Start afterwards
	pool.Start(context.Background())

	// Dispatch must be rejected
	event := model.TaskEvent{
		Type:      model.EventTypeCreated,
		Task:      model.Task{ID: "after-stop"},
		Timestamp: time.Now(),
	}
	if pool.Dispatch(event) {
		t.Error("Dispatch must return false when Stop was called before Start")
	}

	// Another Stop call should also be safe
	pool.Stop()
}

// TestAdversarial_WorkerPool_HighThroughputGracefulShutdown pushes 10,000 events
// across multiple workers, ensuring no leaks, races, or missed events during shutdown.
func TestAdversarial_WorkerPool_HighThroughputGracefulShutdown(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	pool := worker.NewWorkerPool(16, 256, discardLogger)

	pool.Start(context.Background())

	const totalEvents = 10000
	var enqueuedCount atomic.Int64
	var wg sync.WaitGroup

	const numSenders = 20
	const perSender = totalEvents / numSenders

	wg.Add(numSenders)
	for s := 0; s < numSenders; s++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perSender; i++ {
				event := model.TaskEvent{
					Type:      model.EventTypeCreated,
					Task:      model.Task{ID: "stress-event"},
					Timestamp: time.Now(),
				}
				if pool.Dispatch(event) {
					enqueuedCount.Add(1)
				}
			}
		}()
	}

	wg.Wait()
	pool.Stop()

	t.Logf("High throughput graceful shutdown completed: %d/%d events enqueued",
		enqueuedCount.Load(), totalEvents)
}
