package repository

import (
	"context"
	"sort"
	"strings"
	"sync"

	"go-basic-roadmap/internal/model"
)

// MemoryRepository provides a thread-safe, in-memory implementation of TaskRepository.
type MemoryRepository struct {
	mu    sync.RWMutex
	tasks map[string]*model.Task
}

// NewMemoryRepository initializes and returns a new MemoryRepository instance.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tasks: make(map[string]*model.Task),
	}
}

// Create inserts a new task into memory.
func (r *MemoryRepository) Create(ctx context.Context, task *model.Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if task == nil {
		return model.ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; exists {
		return model.ErrTaskConflict
	}

	r.tasks[task.ID] = task.Clone()
	return nil
}

// GetByID retrieves a task by ID from memory, returning a defensive copy.
func (r *MemoryRepository) GetByID(ctx context.Context, id string) (*model.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, model.ErrTaskNotFound
	}

	return task.Clone(), nil
}

// List returns all tasks matching the filter, ordered by CreatedAt descending.
func (r *MemoryRepository) List(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]*model.Task, 0)
	searchTerm := strings.ToLower(strings.TrimSpace(filter.Search))

	for _, task := range r.tasks {
		if filter.Status != "" && task.Status != filter.Status {
			continue
		}
		if filter.Priority != "" && task.Priority != filter.Priority {
			continue
		}
		if searchTerm != "" {
			titleMatch := strings.Contains(strings.ToLower(task.Title), searchTerm)
			descMatch := strings.Contains(strings.ToLower(task.Description), searchTerm)
			if !titleMatch && !descMatch {
				continue
			}
		}
		results = append(results, task.Clone())
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	return results, nil
}

// Update replaces an existing task in memory.
func (r *MemoryRepository) Update(ctx context.Context, task *model.Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if task == nil {
		return model.ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return model.ErrTaskNotFound
	}

	r.tasks[task.ID] = task.Clone()
	return nil
}

// Delete removes a task from memory.
func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return model.ErrTaskNotFound
	}

	delete(r.tasks, id)
	return nil
}
