package repository_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/repository"
)

func TestMemoryRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully creates a new task", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		task := &model.Task{
			ID:          "task-1",
			Title:       "Test Task 1",
			Description: "Description 1",
			Status:      model.StatusTodo,
			Priority:    model.PriorityHigh,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		err := repo.Create(ctx, task)
		if err != nil {
			t.Fatalf("unexpected error creating task: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, "task-1")
		if err != nil {
			t.Fatalf("unexpected error getting created task: %v", err)
		}
		if retrieved.Title != task.Title {
			t.Errorf("got title %q, want %q", retrieved.Title, task.Title)
		}
	})

	t.Run("returns ErrInvalidInput when task is nil", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		err := repo.Create(ctx, nil)
		if !errors.Is(err, model.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("returns ErrTaskConflict on duplicate ID", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		task := &model.Task{
			ID:        "duplicate-id",
			Title:     "First Task",
			Status:    model.StatusTodo,
			Priority:  model.PriorityMedium,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("unexpected error on first create: %v", err)
		}

		duplicate := &model.Task{
			ID:        "duplicate-id",
			Title:     "Second Task with same ID",
			Status:    model.StatusDone,
			Priority:  model.PriorityLow,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := repo.Create(ctx, duplicate)
		if !errors.Is(err, model.ErrTaskConflict) {
			t.Errorf("expected ErrTaskConflict on duplicate ID, got %v", err)
		}
	})

	t.Run("returns context error when context is canceled", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		task := &model.Task{ID: "canceled-task", Title: "Title"}
		err := repo.Create(canceledCtx, task)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})
}

func TestMemoryRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("returns ErrTaskNotFound for non-existent ID", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		_, err := repo.GetByID(ctx, "non-existent")
		if !errors.Is(err, model.ErrTaskNotFound) {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})

	t.Run("returns defensive copy of task", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		now := time.Now().UTC()
		original := &model.Task{
			ID:          "task-defensive",
			Title:       "Original Title",
			Description: "Original Description",
			Status:      model.StatusTodo,
			Priority:    model.PriorityLow,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := repo.Create(ctx, original); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, "task-defensive")
		if err != nil {
			t.Fatalf("failed to get task: %v", err)
		}

		// Mutate the retrieved copy
		retrieved.Title = "Mutated In Caller"
		retrieved.Status = model.StatusDone

		// Retrieve again to verify repository state was not mutated
		second, err := repo.GetByID(ctx, "task-defensive")
		if err != nil {
			t.Fatalf("failed to get task second time: %v", err)
		}

		if second.Title != "Original Title" {
			t.Errorf("repository state was mutated: got %q, want 'Original Title'", second.Title)
		}
		if second.Status != model.StatusTodo {
			t.Errorf("repository status was mutated: got %q, want 'TODO'", second.Status)
		}
	})

	t.Run("returns context error when context is canceled", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := repo.GetByID(canceledCtx, "any-id")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})
}

func TestMemoryRepository_List(t *testing.T) {
	ctx := context.Background()

	setupRepoWithTasks := func() *repository.MemoryRepository {
		repo := repository.NewMemoryRepository()
		baseTime := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

		tasks := []*model.Task{
			{
				ID:          "1",
				Title:       "Learn Go Concurrency",
				Description: "Channels and Goroutines deep dive",
				Status:      model.StatusTodo,
				Priority:    model.PriorityHigh,
				CreatedAt:   baseTime.Add(1 * time.Hour),
				UpdatedAt:   baseTime.Add(1 * time.Hour),
			},
			{
				ID:          "2",
				Title:       "Build REST API",
				Description: "Standard library net/http ServeMux",
				Status:      model.StatusInProgress,
				Priority:    model.PriorityHigh,
				CreatedAt:   baseTime.Add(2 * time.Hour),
				UpdatedAt:   baseTime.Add(2 * time.Hour),
			},
			{
				ID:          "3",
				Title:       "Write Documentation",
				Description: "Roadmap and guide for beginners",
				Status:      model.StatusDone,
				Priority:    model.PriorityMedium,
				CreatedAt:   baseTime.Add(3 * time.Hour),
				UpdatedAt:   baseTime.Add(3 * time.Hour),
			},
			{
				ID:          "4",
				Title:       "Clean up workspace",
				Description: "Tidy files and logs",
				Status:      model.StatusTodo,
				Priority:    model.PriorityLow,
				CreatedAt:   baseTime.Add(4 * time.Hour),
				UpdatedAt:   baseTime.Add(4 * time.Hour),
			},
		}

		for _, task := range tasks {
			if err := repo.Create(ctx, task); err != nil {
				panic(err)
			}
		}
		return repo
	}

	t.Run("empty filter returns all tasks ordered by CreatedAt descending", func(t *testing.T) {
		repo := setupRepoWithTasks()
		results, err := repo.List(ctx, model.TaskFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 4 {
			t.Fatalf("expected 4 tasks, got %d", len(results))
		}

		// Check descending sort order (newest first: 4, 3, 2, 1)
		expectedOrder := []string{"4", "3", "2", "1"}
		for i, id := range expectedOrder {
			if results[i].ID != id {
				t.Errorf("index %d: expected ID %q, got %q", i, id, results[i].ID)
			}
		}
	})

	t.Run("filter by Status", func(t *testing.T) {
		repo := setupRepoWithTasks()
		results, err := repo.List(ctx, model.TaskFilter{
			Status: model.StatusTodo,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("expected 2 tasks with StatusTodo, got %d", len(results))
		}
		for _, task := range results {
			if task.Status != model.StatusTodo {
				t.Errorf("expected StatusTodo, got %s", task.Status)
			}
		}
	})

	t.Run("filter by Priority", func(t *testing.T) {
		repo := setupRepoWithTasks()
		results, err := repo.List(ctx, model.TaskFilter{
			Priority: model.PriorityHigh,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("expected 2 tasks with PriorityHigh, got %d", len(results))
		}
		for _, task := range results {
			if task.Priority != model.PriorityHigh {
				t.Errorf("expected PriorityHigh, got %s", task.Priority)
			}
		}
	})

	t.Run("filter by Search in Title (case-insensitive)", func(t *testing.T) {
		repo := setupRepoWithTasks()
		results, err := repo.List(ctx, model.TaskFilter{
			Search: "concurrency",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 task matching 'concurrency', got %d", len(results))
		}
		if results[0].ID != "1" {
			t.Errorf("expected task ID '1', got %q", results[0].ID)
		}
	})

	t.Run("filter by Search in Description (case-insensitive)", func(t *testing.T) {
		repo := setupRepoWithTasks()
		results, err := repo.List(ctx, model.TaskFilter{
			Search: "servemux",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 task matching description 'servemux', got %d", len(results))
		}
		if results[0].ID != "2" {
			t.Errorf("expected task ID '2', got %q", results[0].ID)
		}
	})

	t.Run("filter with combined criteria", func(t *testing.T) {
		repo := setupRepoWithTasks()
		results, err := repo.List(ctx, model.TaskFilter{
			Status:   model.StatusTodo,
			Priority: model.PriorityHigh,
			Search:   "Channels",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 task matching combined criteria, got %d", len(results))
		}
		if results[0].ID != "1" {
			t.Errorf("expected task ID '1', got %q", results[0].ID)
		}
	})

	t.Run("returns empty slice when no match found", func(t *testing.T) {
		repo := setupRepoWithTasks()
		results, err := repo.List(ctx, model.TaskFilter{
			Search: "nonexistent query terms",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if results == nil {
			t.Error("expected non-nil empty slice, got nil")
		}
		if len(results) != 0 {
			t.Errorf("expected 0 tasks, got %d", len(results))
		}
	})

	t.Run("returns context error when context is canceled", func(t *testing.T) {
		repo := setupRepoWithTasks()
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := repo.List(canceledCtx, model.TaskFilter{})
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})
}

func TestMemoryRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully updates existing task", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		initial := &model.Task{
			ID:          "task-update",
			Title:       "Original Title",
			Description: "Original Desc",
			Status:      model.StatusTodo,
			Priority:    model.PriorityLow,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := repo.Create(ctx, initial); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		updated := &model.Task{
			ID:          "task-update",
			Title:       "New Title",
			Description: "New Desc",
			Status:      model.StatusInProgress,
			Priority:    model.PriorityHigh,
			CreatedAt:   initial.CreatedAt,
			UpdatedAt:   time.Now().UTC(),
		}

		if err := repo.Update(ctx, updated); err != nil {
			t.Fatalf("unexpected error on update: %v", err)
		}

		fetched, err := repo.GetByID(ctx, "task-update")
		if err != nil {
			t.Fatalf("failed to get updated task: %v", err)
		}

		if fetched.Title != "New Title" || fetched.Status != model.StatusInProgress || fetched.Priority != model.PriorityHigh {
			t.Errorf("updated task has unexpected fields: %+v", fetched)
		}
	})

	t.Run("returns ErrInvalidInput on nil task", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		err := repo.Update(ctx, nil)
		if !errors.Is(err, model.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("returns ErrTaskNotFound when updating non-existent task", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		task := &model.Task{ID: "does-not-exist", Title: "Title"}
		err := repo.Update(ctx, task)
		if !errors.Is(err, model.ErrTaskNotFound) {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})

	t.Run("returns context error when context is canceled", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		task := &model.Task{ID: "any-id", Title: "Title"}
		err := repo.Update(canceledCtx, task)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})
}

func TestMemoryRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully deletes existing task", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		task := &model.Task{
			ID:        "to-delete",
			Title:     "Delete Me",
			Status:    model.StatusTodo,
			Priority:  model.PriorityLow,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		if err := repo.Delete(ctx, "to-delete"); err != nil {
			t.Fatalf("unexpected error on delete: %v", err)
		}

		// Subsequent GetByID must return ErrTaskNotFound
		_, err := repo.GetByID(ctx, "to-delete")
		if !errors.Is(err, model.ErrTaskNotFound) {
			t.Errorf("expected ErrTaskNotFound after delete, got %v", err)
		}
	})

	t.Run("returns ErrTaskNotFound when deleting non-existent task", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		err := repo.Delete(ctx, "missing-task")
		if !errors.Is(err, model.ErrTaskNotFound) {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})

	t.Run("returns context error when context is canceled", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		err := repo.Delete(canceledCtx, "any-id")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})
}

func TestMemoryRepository_ConcurrencyRace(t *testing.T) {
	// Concurrency stress test: 60 concurrent goroutines continuously
	// reading, creating, updating, and listing tasks to verify thread safety
	// under sync.RWMutex.
	repo := repository.NewMemoryRepository()
	ctx := context.Background()

	const numGoroutines = 60
	const opsPerGoroutine = 50

	// Seed with initial tasks
	for i := 0; i < 10; i++ {
		task := &model.Task{
			ID:          fmt.Sprintf("seed-%d", i),
			Title:       fmt.Sprintf("Seed Task %d", i),
			Description: "Initial seed",
			Status:      model.StatusTodo,
			Priority:    model.PriorityMedium,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("failed to seed task: %v", err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		goroutineID := g
		go func() {
			defer wg.Done()

			for op := 0; op < opsPerGoroutine; op++ {
				switch op % 5 {
				case 0:
					// Create new task
					task := &model.Task{
						ID:          fmt.Sprintf("concurrent-%d-%d", goroutineID, op),
						Title:       fmt.Sprintf("Concurrent Task %d-%d", goroutineID, op),
						Description: "Stress test task",
						Status:      model.StatusTodo,
						Priority:    model.PriorityHigh,
						CreatedAt:   time.Now().UTC(),
						UpdatedAt:   time.Now().UTC(),
					}
					_ = repo.Create(ctx, task)

				case 1:
					// Read existing task
					targetID := fmt.Sprintf("seed-%d", op%10)
					_, _ = repo.GetByID(ctx, targetID)

				case 2:
					// Update existing task
					targetID := fmt.Sprintf("seed-%d", op%10)
					updated := &model.Task{
						ID:          targetID,
						Title:       fmt.Sprintf("Updated Seed %d by goroutine %d", op%10, goroutineID),
						Description: "Updated desc",
						Status:      model.StatusInProgress,
						Priority:    model.PriorityHigh,
						CreatedAt:   time.Now().UTC(),
						UpdatedAt:   time.Now().UTC(),
					}
					_ = repo.Update(ctx, updated)

				case 3:
					// List with filters
					_, _ = repo.List(ctx, model.TaskFilter{
						Status: model.StatusTodo,
					})

				case 4:
					// Delete created task
					targetID := fmt.Sprintf("concurrent-%d-%d", goroutineID, op-4)
					_ = repo.Delete(ctx, targetID)
				}
			}
		}()
	}

	wg.Wait()

	// Final verification: ensure repository remains in a consistent state
	tasks, err := repo.List(ctx, model.TaskFilter{})
	if err != nil {
		t.Fatalf("failed to list tasks after concurrency test: %v", err)
	}

	if len(tasks) < 10 {
		t.Errorf("expected at least 10 tasks remaining, got %d", len(tasks))
	}
}
