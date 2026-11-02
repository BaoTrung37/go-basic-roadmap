package service_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/repository"
	"go-basic-roadmap/internal/service"
)

// mockTaskRepository implements repository.TaskRepository for isolated unit testing.
type mockTaskRepository struct {
	createFunc  func(ctx context.Context, task *model.Task) error
	getByIDFunc func(ctx context.Context, id string) (*model.Task, error)
	listFunc    func(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error)
	updateFunc  func(ctx context.Context, task *model.Task) error
	deleteFunc  func(ctx context.Context, id string) error
}

func (m *mockTaskRepository) Create(ctx context.Context, task *model.Task) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, task)
	}
	return nil
}

func (m *mockTaskRepository) GetByID(ctx context.Context, id string) (*model.Task, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, model.ErrTaskNotFound
}

func (m *mockTaskRepository) List(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockTaskRepository) Update(ctx context.Context, task *model.Task) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, task)
	}
	return nil
}

func (m *mockTaskRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// Ensure mockTaskRepository satisfies repository.TaskRepository at compile time.
var _ repository.TaskRepository = (*mockTaskRepository)(nil)

// mockWorkerPool implements worker.WorkerPool and safely captures emitted events.
type mockWorkerPool struct {
	mu         sync.Mutex
	dispatched []model.TaskEvent
}

func (m *mockWorkerPool) Start(ctx context.Context) {}

func (m *mockWorkerPool) Dispatch(event model.TaskEvent) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dispatched = append(m.dispatched, event)
	return true
}

func (m *mockWorkerPool) Stop() {}

func (m *mockWorkerPool) Events() []model.TaskEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]model.TaskEvent, len(m.dispatched))
	copy(copied, m.dispatched)
	return copied
}

func TestNewTaskService(t *testing.T) {
	t.Run("constructs with default logger when nil logger passed", func(t *testing.T) {
		repo := &mockTaskRepository{}
		pool := &mockWorkerPool{}
		svc := service.NewTaskService(repo, pool, nil)
		if svc == nil {
			t.Fatal("expected non-nil TaskService")
		}
	})

	t.Run("functions normally when pool is nil", func(t *testing.T) {
		repo := &mockTaskRepository{
			createFunc: func(ctx context.Context, task *model.Task) error {
				return nil
			},
		}
		svc := service.NewTaskService(repo, nil, nil)
		task, err := svc.CreateTask(context.Background(), "No Pool Task", "Description", model.PriorityLow)
		if err != nil {
			t.Fatalf("unexpected error when pool is nil: %v", err)
		}
		if task == nil {
			t.Fatal("expected non-nil task returned")
		}
	})
}

func TestTaskService_CreateTask(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		title         string
		desc          string
		priority      model.Priority
		repoError     error
		wantErr       bool
		expectedErrIs error
		checkEvent    bool
	}{
		{
			name:       "success with explicit priority",
			title:      "Build Task Service",
			desc:       "Write unit tests for task service",
			priority:   model.PriorityHigh,
			wantErr:    false,
			checkEvent: true,
		},
		{
			name:       "success with default priority when empty",
			title:      "Default Priority Task",
			desc:       "Should default to MEDIUM",
			priority:   "",
			wantErr:    false,
			checkEvent: true,
		},
		{
			name:          "error: empty title fails validation",
			title:         "",
			desc:          "Some description",
			priority:      model.PriorityMedium,
			wantErr:       true,
			expectedErrIs: model.ErrInvalidInput,
			checkEvent:    false,
		},
		{
			name:          "error: whitespace title fails validation",
			title:         "   \t  ",
			desc:          "Some description",
			priority:      model.PriorityMedium,
			wantErr:       true,
			expectedErrIs: model.ErrInvalidInput,
			checkEvent:    false,
		},
		{
			name:          "error: invalid priority fails validation",
			title:         "Valid Title",
			desc:          "Desc",
			priority:      model.Priority("NON_EXISTENT"),
			wantErr:       true,
			expectedErrIs: model.ErrInvalidInput,
			checkEvent:    false,
		},
		{
			name:          "error: repository returns storage failure",
			title:         "Valid Title",
			desc:          "Desc",
			priority:      model.PriorityLow,
			repoError:     model.ErrStorageFailure,
			wantErr:       true,
			expectedErrIs: model.ErrStorageFailure,
			checkEvent:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pool := &mockWorkerPool{}
			repo := &mockTaskRepository{
				createFunc: func(ctx context.Context, task *model.Task) error {
					if tc.repoError != nil {
						return tc.repoError
					}
					return nil
				},
			}

			svc := service.NewTaskService(repo, pool, nil)
			created, err := svc.CreateTask(ctx, tc.title, tc.desc, tc.priority)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.expectedErrIs != nil && !errors.Is(err, tc.expectedErrIs) {
					t.Errorf("expected error wrapping %v, got %v", tc.expectedErrIs, err)
				}
				if len(pool.Events()) > 0 {
					t.Errorf("unexpected event dispatched on error: %v", pool.Events())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if created == nil {
					t.Fatal("expected non-nil created task")
				}
				if created.Title != strings.TrimSpace(tc.title) {
					t.Errorf("got title %q, want %q", created.Title, tc.title)
				}
				if created.Status != model.StatusTodo {
					t.Errorf("expected initial status TODO, got %s", created.Status)
				}
				if tc.priority == "" {
					if created.Priority != model.PriorityMedium {
						t.Errorf("expected default priority MEDIUM, got %s", created.Priority)
					}
				} else {
					if created.Priority != tc.priority {
						t.Errorf("got priority %s, want %s", created.Priority, tc.priority)
					}
				}

				if tc.checkEvent {
					events := pool.Events()
					if len(events) != 1 {
						t.Fatalf("expected 1 event dispatched, got %d", len(events))
					}
					if events[0].Type != model.EventTypeCreated {
						t.Errorf("expected event type CREATED, got %q", events[0].Type)
					}
					if events[0].Task.ID != created.ID {
						t.Errorf("event task ID %q mismatch with created ID %q", events[0].Task.ID, created.ID)
					}
				}
			}
		})
	}
}

func TestTaskService_GetTask(t *testing.T) {
	ctx := context.Background()

	t.Run("returns ErrInvalidInput on empty or whitespace ID", func(t *testing.T) {
		svc := service.NewTaskService(&mockTaskRepository{}, nil, nil)
		for _, emptyID := range []string{"", "   ", "\t\n"} {
			_, err := svc.GetTask(ctx, emptyID)
			if !errors.Is(err, model.ErrInvalidInput) {
				t.Errorf("GetTask(%q) expected ErrInvalidInput, got %v", emptyID, err)
			}
		}
	})

	t.Run("success returns task from repository", func(t *testing.T) {
		expected := &model.Task{
			ID:    "task-456",
			Title: "Test Task",
		}
		repo := &mockTaskRepository{
			getByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				if id == "task-456" {
					return expected, nil
				}
				return nil, model.ErrTaskNotFound
			},
		}

		svc := service.NewTaskService(repo, nil, nil)
		got, err := svc.GetTask(ctx, "task-456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != expected.ID || got.Title != expected.Title {
			t.Errorf("got %+v, want %+v", got, expected)
		}
	})

	t.Run("propagates ErrTaskNotFound when task missing", func(t *testing.T) {
		repo := &mockTaskRepository{
			getByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return nil, model.ErrTaskNotFound
			},
		}
		svc := service.NewTaskService(repo, nil, nil)
		_, err := svc.GetTask(ctx, "missing-id")
		if !errors.Is(err, model.ErrTaskNotFound) {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})
}

func TestTaskService_ListTasks(t *testing.T) {
	ctx := context.Background()

	t.Run("delegates to repository List with filter", func(t *testing.T) {
		expectedTasks := []*model.Task{
			{ID: "1", Title: "Task 1", Status: model.StatusTodo},
			{ID: "2", Title: "Task 2", Status: model.StatusTodo},
		}

		filterGiven := model.TaskFilter{
			Status: model.StatusTodo,
		}

		var receivedFilter model.TaskFilter
		repo := &mockTaskRepository{
			listFunc: func(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
				receivedFilter = filter
				return expectedTasks, nil
			},
		}

		svc := service.NewTaskService(repo, nil, nil)
		results, err := svc.ListTasks(ctx, filterGiven)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != len(expectedTasks) {
			t.Fatalf("got %d tasks, want %d", len(results), len(expectedTasks))
		}
		if receivedFilter.Status != model.StatusTodo {
			t.Errorf("repository received filter %v, want %v", receivedFilter, filterGiven)
		}
	})

	t.Run("propagates repository list error", func(t *testing.T) {
		repo := &mockTaskRepository{
			listFunc: func(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
				return nil, model.ErrStorageFailure
			},
		}
		svc := service.NewTaskService(repo, nil, nil)
		_, err := svc.ListTasks(ctx, model.TaskFilter{})
		if !errors.Is(err, model.ErrStorageFailure) {
			t.Errorf("expected ErrStorageFailure, got %v", err)
		}
	})
}

func TestTaskService_UpdateTask(t *testing.T) {
	ctx := context.Background()

	t.Run("returns ErrInvalidInput on empty or whitespace ID", func(t *testing.T) {
		svc := service.NewTaskService(&mockTaskRepository{}, nil, nil)
		_, err := svc.UpdateTask(ctx, "  ", "Title", "Desc", model.StatusInProgress, model.PriorityHigh)
		if !errors.Is(err, model.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("returns validation error on invalid input fields", func(t *testing.T) {
		svc := service.NewTaskService(&mockTaskRepository{}, nil, nil)
		_, err := svc.UpdateTask(ctx, "id-1", "", "Desc", model.StatusInProgress, model.PriorityHigh)
		if !errors.Is(err, model.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput on empty title, got %v", err)
		}

		_, err = svc.UpdateTask(ctx, "id-1", "Title", "Desc", model.Status("INVALID"), model.PriorityHigh)
		if !errors.Is(err, model.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput on invalid status, got %v", err)
		}
	})

	t.Run("returns ErrTaskNotFound when task does not exist in repo", func(t *testing.T) {
		repo := &mockTaskRepository{
			getByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return nil, model.ErrTaskNotFound
			},
		}
		svc := service.NewTaskService(repo, nil, nil)
		_, err := svc.UpdateTask(ctx, "id-missing", "Title", "Desc", model.StatusInProgress, model.PriorityHigh)
		if !errors.Is(err, model.ErrTaskNotFound) {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})

	t.Run("emits UPDATED event when status is unchanged", func(t *testing.T) {
		pool := &mockWorkerPool{}
		existing := &model.Task{
			ID:          "task-same-status",
			Title:       "Old Title",
			Description: "Old Desc",
			Status:      model.StatusInProgress,
			Priority:    model.PriorityLow,
			CreatedAt:   time.Now().UTC().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().UTC().Add(-1 * time.Hour),
		}

		var updatedPersisted *model.Task
		repo := &mockTaskRepository{
			getByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return existing, nil
			},
			updateFunc: func(ctx context.Context, task *model.Task) error {
				updatedPersisted = task
				return nil
			},
		}

		svc := service.NewTaskService(repo, pool, nil)
		updated, err := svc.UpdateTask(ctx, "task-same-status", "New Title", "New Desc", model.StatusInProgress, model.PriorityHigh)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updated.Title != "New Title" || updated.Priority != model.PriorityHigh {
			t.Errorf("unexpected updated task state: %+v", updated)
		}
		if updatedPersisted == nil || updatedPersisted.Title != "New Title" {
			t.Errorf("task was not properly persisted: %+v", updatedPersisted)
		}

		events := pool.Events()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].Type != model.EventTypeUpdated {
			t.Errorf("expected event type UPDATED, got %q", events[0].Type)
		}
	})

	t.Run("emits STATUS_CHANGED event when status is modified", func(t *testing.T) {
		pool := &mockWorkerPool{}
		existing := &model.Task{
			ID:          "task-diff-status",
			Title:       "Task",
			Description: "Desc",
			Status:      model.StatusTodo,
			Priority:    model.PriorityLow,
			CreatedAt:   time.Now().UTC().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().UTC().Add(-1 * time.Hour),
		}

		repo := &mockTaskRepository{
			getByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return existing, nil
			},
			updateFunc: func(ctx context.Context, task *model.Task) error {
				return nil
			},
		}

		svc := service.NewTaskService(repo, pool, nil)
		_, err := svc.UpdateTask(ctx, "task-diff-status", "Task", "Desc", model.StatusDone, model.PriorityLow)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		events := pool.Events()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].Type != model.EventTypeStatusChanged {
			t.Errorf("expected event type STATUS_CHANGED, got %q", events[0].Type)
		}
	})

	t.Run("propagates repository update error", func(t *testing.T) {
		existing := &model.Task{
			ID:     "task-1",
			Title:  "Old Title",
			Status: model.StatusTodo,
		}
		repo := &mockTaskRepository{
			getByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return existing, nil
			},
			updateFunc: func(ctx context.Context, task *model.Task) error {
				return model.ErrStorageFailure
			},
		}

		svc := service.NewTaskService(repo, nil, nil)
		_, err := svc.UpdateTask(ctx, "task-1", "New Title", "Desc", model.StatusTodo, model.PriorityMedium)
		if !errors.Is(err, model.ErrStorageFailure) {
			t.Errorf("expected ErrStorageFailure, got %v", err)
		}
	})
}

func TestTaskService_DeleteTask(t *testing.T) {
	ctx := context.Background()

	t.Run("returns ErrInvalidInput on empty or whitespace ID", func(t *testing.T) {
		svc := service.NewTaskService(&mockTaskRepository{}, nil, nil)
		err := svc.DeleteTask(ctx, "   ")
		if !errors.Is(err, model.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("returns ErrTaskNotFound when task does not exist", func(t *testing.T) {
		repo := &mockTaskRepository{
			getByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return nil, model.ErrTaskNotFound
			},
		}
		svc := service.NewTaskService(repo, nil, nil)
		err := svc.DeleteTask(ctx, "missing-task")
		if !errors.Is(err, model.ErrTaskNotFound) {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})

	t.Run("successfully deletes task and emits DELETED event", func(t *testing.T) {
		pool := &mockWorkerPool{}
		existing := &model.Task{
			ID:        "to-delete",
			Title:     "Task to be deleted",
			Status:    model.StatusTodo,
			Priority:  model.PriorityLow,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		deletedID := ""
		repo := &mockTaskRepository{
			getByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return existing, nil
			},
			deleteFunc: func(ctx context.Context, id string) error {
				deletedID = id
				return nil
			},
		}

		svc := service.NewTaskService(repo, pool, nil)
		err := svc.DeleteTask(ctx, "to-delete")
		if err != nil {
			t.Fatalf("unexpected error on delete: %v", err)
		}

		if deletedID != "to-delete" {
			t.Errorf("expected repo.Delete called with 'to-delete', got %q", deletedID)
		}

		events := pool.Events()
		if len(events) != 1 {
			t.Fatalf("expected 1 event dispatched, got %d", len(events))
		}
		if events[0].Type != model.EventTypeDeleted {
			t.Errorf("expected event type DELETED, got %q", events[0].Type)
		}
		if events[0].Task.ID != "to-delete" {
			t.Errorf("expected event task ID 'to-delete', got %q", events[0].Task.ID)
		}
	})

	t.Run("propagates repository delete error without emitting event", func(t *testing.T) {
		pool := &mockWorkerPool{}
		existing := &model.Task{ID: "task-fail-del"}
		repo := &mockTaskRepository{
			getByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return existing, nil
			},
			deleteFunc: func(ctx context.Context, id string) error {
				return model.ErrStorageFailure
			},
		}

		svc := service.NewTaskService(repo, pool, nil)
		err := svc.DeleteTask(ctx, "task-fail-del")
		if !errors.Is(err, model.ErrStorageFailure) {
			t.Errorf("expected ErrStorageFailure, got %v", err)
		}

		if len(pool.Events()) > 0 {
			t.Errorf("unexpected event dispatched on delete failure")
		}
	})
}
