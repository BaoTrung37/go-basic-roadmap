package worker_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/worker"
)

// discardLogger creates a logger writing to io.Discard to keep test output clean.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewWorkerPool(t *testing.T) {
	t.Run("applies defaults for non-positive parameters and nil logger", func(t *testing.T) {
		pool := worker.NewWorkerPool(0, 0, nil)
		if pool == nil {
			t.Fatal("expected non-nil WorkerPool")
		}
		if pool.Capacity() != 10 {
			t.Errorf("expected default buffer capacity 10, got %d", pool.Capacity())
		}
		if pool.QueueLength() != 0 {
			t.Errorf("expected initial queue length 0, got %d", pool.QueueLength())
		}
	})

	t.Run("respects custom parameters", func(t *testing.T) {
		pool := worker.NewWorkerPool(5, 50, discardLogger())
		if pool.Capacity() != 50 {
			t.Errorf("expected capacity 50, got %d", pool.Capacity())
		}
	})
}

func TestWorkerPool_DispatchPreconditions(t *testing.T) {
	t.Run("dispatch before start returns false", func(t *testing.T) {
		pool := worker.NewWorkerPool(2, 10, discardLogger())
		event := model.TaskEvent{
			Type:      model.EventTypeCreated,
			Task:      model.Task{ID: "task-1"},
			Timestamp: time.Now().UTC(),
		}

		ok := pool.Dispatch(event)
		if ok {
			t.Error("expected Dispatch to return false before Start(), got true")
		}
	})

	t.Run("dispatch after stop returns false", func(t *testing.T) {
		pool := worker.NewWorkerPool(2, 10, discardLogger())
		ctx := context.Background()
		pool.Start(ctx)
		pool.Stop()

		event := model.TaskEvent{
			Type:      model.EventTypeCreated,
			Task:      model.Task{ID: "task-2"},
			Timestamp: time.Now().UTC(),
		}

		ok := pool.Dispatch(event)
		if ok {
			t.Error("expected Dispatch to return false after Stop(), got true")
		}
	})
}

func TestWorkerPool_BufferFullBehavior(t *testing.T) {
	// Test that when the buffer fills up, Dispatch drops event and returns false without blocking
	const bufCap = 3
	pool := worker.NewWorkerPool(1, bufCap, discardLogger())

	// We start the pool but block worker consumption by filling queue before starting or pausing
	// Instead of blocking worker, we can test non-blocking nature by saturating a small queue:
	// Fill queue to capacity while pool has not drained it yet
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool.Start(ctx)

	// Dispatch up to capacity rapidly
	dispatchedCount := 0
	droppedCount := 0
	for i := 0; i < 50; i++ {
		event := model.TaskEvent{
			Type:      model.EventTypeCreated,
			Task:      model.Task{ID: fmt.Sprintf("saturation-%d", i)},
			Timestamp: time.Now().UTC(),
		}
		if pool.Dispatch(event) {
			dispatchedCount++
		} else {
			droppedCount++
		}
	}

	// Because queue capacity is 3 and worker is processing asynchronously,
	// some items must have been dispatched. The operation MUST complete promptly without blocking.
	if dispatchedCount == 0 {
		t.Error("expected at least some events to be successfully dispatched")
	}

	pool.Stop()
}

func TestWorkerPool_EventProcessingAndGracefulShutdown(t *testing.T) {
	pool := worker.NewWorkerPool(3, 20, discardLogger())
	ctx := context.Background()

	pool.Start(ctx)

	// Calling Start multiple times should be a no-op
	pool.Start(ctx)

	const eventCount = 15
	for i := 0; i < eventCount; i++ {
		event := model.TaskEvent{
			Type: model.EventTypeCreated,
			Task: model.Task{
				ID:       fmt.Sprintf("task-%d", i),
				Title:    fmt.Sprintf("Title %d", i),
				Status:   model.StatusTodo,
				Priority: model.PriorityMedium,
			},
			Timestamp: time.Now().UTC(),
		}

		dispatched := pool.Dispatch(event)
		if !dispatched {
			t.Fatalf("failed to dispatch event %d", i)
		}
	}

	// Stop must drain all buffered items and shut down without deadlocking
	stopDone := make(chan struct{})
	go func() {
		pool.Stop()
		close(stopDone)
	}()

	select {
	case <-stopDone:
		// Success: workers exited cleanly
	case <-time.After(3 * time.Second):
		t.Fatal("worker pool Stop() timed out: potential deadlock during shutdown")
	}

	// Calling Stop multiple times should be safe (sync.Once)
	pool.Stop()
}

func TestWorkerPool_ParentContextCancellation(t *testing.T) {
	pool := worker.NewWorkerPool(2, 10, discardLogger())
	ctx, cancel := context.WithCancel(context.Background())

	pool.Start(ctx)

	// Dispatch a few events
	for i := 0; i < 5; i++ {
		pool.Dispatch(model.TaskEvent{
			Type:      model.EventTypeCreated,
			Task:      model.Task{ID: fmt.Sprintf("ctx-task-%d", i)},
			Timestamp: time.Now().UTC(),
		})
	}

	// Cancel parent context
	cancel()

	// Stop pool; it should drain and terminate promptly
	stopDone := make(chan struct{})
	go func() {
		pool.Stop()
		close(stopDone)
	}()

	select {
	case <-stopDone:
		// Clean exit
	case <-time.After(3 * time.Second):
		t.Fatal("worker pool failed to terminate after context cancellation")
	}
}

func TestWorkerPool_ConcurrentDispatchStress(t *testing.T) {
	// Concurrency test: multiple goroutines continuously dispatching events
	pool := worker.NewWorkerPool(4, 50, discardLogger())
	ctx := context.Background()
	pool.Start(ctx)

	const numDispatchers = 20
	const eventsPerDispatcher = 30

	var wg sync.WaitGroup
	wg.Add(numDispatchers)

	for d := 0; d < numDispatchers; d++ {
		dispatcherID := d
		go func() {
			defer wg.Done()
			for e := 0; e < eventsPerDispatcher; e++ {
				event := model.TaskEvent{
					Type: model.EventTypeUpdated,
					Task: model.Task{
						ID:       fmt.Sprintf("concurrent-%d-%d", dispatcherID, e),
						Status:   model.StatusInProgress,
						Priority: model.PriorityHigh,
					},
					Timestamp: time.Now().UTC(),
				}
				pool.Dispatch(event)
				time.Sleep(100 * time.Microsecond)
			}
		}()
	}

	wg.Wait()

	stopDone := make(chan struct{})
	go func() {
		pool.Stop()
		close(stopDone)
	}()

	select {
	case <-stopDone:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("worker pool Stop() timed out during concurrent stress shutdown")
	}
}
