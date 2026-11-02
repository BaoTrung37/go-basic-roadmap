package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/repository"
	"go-basic-roadmap/internal/worker"
)

// TaskService orchestrates business workflows for managing tasks.
type TaskService interface {
	CreateTask(ctx context.Context, title, description string, priority model.Priority) (*model.Task, error)
	GetTask(ctx context.Context, id string) (*model.Task, error)
	ListTasks(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error)
	UpdateTask(ctx context.Context, id, title, description string, status model.Status, priority model.Priority) (*model.Task, error)
	DeleteTask(ctx context.Context, id string) error
}

// taskService implements the TaskService interface.
type taskService struct {
	repo   repository.TaskRepository
	pool   worker.WorkerPool
	logger *slog.Logger
}

// NewTaskService constructs a new TaskService instance.
func NewTaskService(repo repository.TaskRepository, pool worker.WorkerPool, logger *slog.Logger) TaskService {
	if logger == nil {
		logger = slog.Default()
	}
	return &taskService{
		repo:   repo,
		pool:   pool,
		logger: logger,
	}
}

// CreateTask validates input, creates a new task entity, stores it, and emits a CREATED event.
func (s *taskService) CreateTask(ctx context.Context, title, description string, priority model.Priority) (*model.Task, error) {
	if priority == "" {
		priority = model.PriorityMedium
	}

	if err := model.ValidateCreateTask(title, priority); err != nil {
		return nil, err
	}

	id, err := model.GenerateID()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate task ID: %v", model.ErrStorageFailure, err)
	}

	now := time.Now().UTC()
	task := &model.Task{
		ID:          id,
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		Status:      model.StatusTodo,
		Priority:    priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	if s.pool != nil {
		s.pool.Dispatch(model.TaskEvent{
			Type:      model.EventTypeCreated,
			Task:      *task,
			Timestamp: now,
		})
	}

	return task, nil
}

// GetTask retrieves a task by ID.
func (s *taskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, fmt.Errorf("%w: task ID cannot be empty", model.ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, cleanID)
}

// ListTasks returns tasks matching the provided filter criteria.
func (s *taskService) ListTasks(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
	return s.repo.List(ctx, filter)
}

// UpdateTask validates inputs, checks for task existence, updates fields, and dispatches events.
func (s *taskService) UpdateTask(ctx context.Context, id, title, description string, status model.Status, priority model.Priority) (*model.Task, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, fmt.Errorf("%w: task ID cannot be empty", model.ErrInvalidInput)
	}

	if err := model.ValidateUpdateTask(title, status, priority); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, cleanID)
	if err != nil {
		return nil, err
	}

	statusChanged := (existing.Status != status)
	now := time.Now().UTC()

	updated := &model.Task{
		ID:          cleanID,
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		Status:      status,
		Priority:    priority,
		CreatedAt:   existing.CreatedAt,
		UpdatedAt:   now,
	}

	if err := s.repo.Update(ctx, updated); err != nil {
		return nil, err
	}

	if s.pool != nil {
		eventType := model.EventTypeUpdated
		if statusChanged {
			eventType = model.EventTypeStatusChanged
		}
		s.pool.Dispatch(model.TaskEvent{
			Type:      eventType,
			Task:      *updated,
			Timestamp: now,
		})
	}

	return updated, nil
}

// DeleteTask verifies task existence, deletes the task from storage, and emits a DELETED event.
func (s *taskService) DeleteTask(ctx context.Context, id string) error {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return fmt.Errorf("%w: task ID cannot be empty", model.ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, cleanID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, cleanID); err != nil {
		return err
	}

	if s.pool != nil {
		s.pool.Dispatch(model.TaskEvent{
			Type:      model.EventTypeDeleted,
			Task:      *existing,
			Timestamp: time.Now().UTC(),
		})
	}

	return nil
}
