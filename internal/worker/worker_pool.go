package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"go-basic-roadmap/internal/model"
)

// WorkerPool defines the contract for an asynchronous background task processor.
type WorkerPool interface {
	Start(ctx context.Context)
	Dispatch(event model.TaskEvent) bool
	Stop()
}

// Pool implements WorkerPool using goroutines, buffered channels, and sync.WaitGroup.
type Pool struct {
	mu         sync.RWMutex
	numWorkers int
	capacity   int
	queue      chan model.TaskEvent
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *slog.Logger
	started    bool
	stopped    bool
	stopOnce   sync.Once
}

// NewWorkerPool constructs a new Pool with the specified number of workers and buffer size.
func NewWorkerPool(numWorkers int, bufferCapacity int, logger *slog.Logger) *Pool {
	if numWorkers <= 0 {
		numWorkers = 1
	}
	if bufferCapacity <= 0 {
		bufferCapacity = 10
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Pool{
		numWorkers: numWorkers,
		capacity:   bufferCapacity,
		queue:      make(chan model.TaskEvent, bufferCapacity),
		logger:     logger,
	}
}

// Start launches worker goroutines that process events from the internal buffered channel.
func (p *Pool) Start(ctx context.Context) {
	p.mu.Lock()
	if p.started || p.stopped {
		p.mu.Unlock()
		return
	}
	p.started = true
	if ctx == nil {
		ctx = context.Background()
	}
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.wg.Add(p.numWorkers)
	p.mu.Unlock()

	p.logger.Info("Starting worker pool",
		"workers", p.numWorkers,
		"buffer_capacity", p.capacity,
	)

	for i := 1; i <= p.numWorkers; i++ {
		go p.worker(i)
	}
}

// Dispatch enqueues an event for asynchronous processing.
// Returns true if enqueued successfully, or false if the queue is full or stopped.
func (p *Pool) Dispatch(event model.TaskEvent) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started || p.stopped {
		return false
	}

	select {
	case p.queue <- event:
		return true
	default:
		p.logger.Warn("Worker event queue is full, dropping event",
			"event_type", event.Type,
			"task_id", event.Task.ID,
		)
		return false
	}
}

// Stop initiates graceful shutdown: marks the pool stopped, closes the channel,
// drains pending items, and waits for all worker goroutines to terminate.
func (p *Pool) Stop() {
	p.stopOnce.Do(func() {
		p.logger.Info("Stopping worker pool...")

		p.mu.Lock()
		p.stopped = true
		close(p.queue)
		p.mu.Unlock()

		p.wg.Wait()

		if p.cancel != nil {
			p.cancel()
		}

		p.logger.Info("Worker pool stopped cleanly")
	})
}

// worker is the worker goroutine loop.
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			// Context canceled: drain any items remaining in the closed or pending queue
			for {
				select {
				case event, ok := <-p.queue:
					if !ok {
						return
					}
					p.processEvent(id, event)
				default:
					return
				}
			}
		case event, ok := <-p.queue:
			if !ok {
				return
			}
			p.processEvent(id, event)
		}
	}
}

// processEvent handles an individual TaskEvent (logging, audit recording).
func (p *Pool) processEvent(workerID int, event model.TaskEvent) {
	p.logger.Info("Worker processed task event",
		"worker_id", workerID,
		"event_type", event.Type,
		"task_id", event.Task.ID,
		"status", event.Task.Status,
		"priority", event.Task.Priority,
		"timestamp", event.Timestamp.Format(time.RFC3339),
	)
}

// QueueLength returns the current number of queued, unhandled events.
func (p *Pool) QueueLength() int {
	return len(p.queue)
}

// Capacity returns the total buffer capacity of the event queue.
func (p *Pool) Capacity() int {
	return p.capacity
}
