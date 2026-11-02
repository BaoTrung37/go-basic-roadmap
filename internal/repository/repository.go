package repository

import (
	"context"
	"go-basic-roadmap/internal/model"
)

// TaskRepository defines the storage abstraction for tasks.
// Decoupling data storage behind this interface allows swapping between
// in-memory, file-backed, or external database engines without changing business logic.
type TaskRepository interface {
	// Create persists a new task. Returns ErrTaskConflict if a task with the ID already exists.
	Create(ctx context.Context, task *model.Task) error

	// GetByID retrieves a task by its unique ID. Returns ErrTaskNotFound if not found.
	GetByID(ctx context.Context, id string) (*model.Task, error)

	// List returns tasks matching the specified filter criteria.
	List(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error)

	// Update updates an existing task. Returns ErrTaskNotFound if the task does not exist.
	Update(ctx context.Context, task *model.Task) error

	// Delete removes a task by ID. Returns ErrTaskNotFound if the task does not exist.
	Delete(ctx context.Context, id string) error
}
