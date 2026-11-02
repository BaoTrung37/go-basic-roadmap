package repository_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/repository"
)

func TestNewFileRepository(t *testing.T) {
	t.Run("creates parent directory and initializes empty file when non-existent", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "nested", "sub", "tasks.json")

		repo, err := repository.NewFileRepository(filePath)
		if err != nil {
			t.Fatalf("unexpected error initializing file repository: %v", err)
		}
		if repo == nil {
			t.Fatal("expected non-nil repository")
		}

		// Verify file was created on disk
		info, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("file does not exist on disk: %v", err)
		}
		if info.Size() == 0 {
			t.Errorf("expected file to have initialized JSON content, size is 0")
		}

		// Verify listing is empty
		tasks, err := repo.List(context.Background(), model.TaskFilter{})
		if err != nil {
			t.Fatalf("unexpected list error: %v", err)
		}
		if len(tasks) != 0 {
			t.Errorf("expected 0 tasks, got %d", len(tasks))
		}
	})

	t.Run("handles existing file with whitespace gracefully", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "whitespace.json")
		if err := os.WriteFile(filePath, []byte("   \n\t  \n"), 0644); err != nil {
			t.Fatalf("failed to create dummy file: %v", err)
		}

		repo, err := repository.NewFileRepository(filePath)
		if err != nil {
			t.Fatalf("unexpected error with whitespace file: %v", err)
		}

		tasks, err := repo.List(context.Background(), model.TaskFilter{})
		if err != nil {
			t.Fatalf("unexpected list error: %v", err)
		}
		if len(tasks) != 0 {
			t.Errorf("expected 0 tasks, got %d", len(tasks))
		}
	})

	t.Run("returns ErrStorageFailure on invalid/corrupted JSON file", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "corrupted.json")
		if err := os.WriteFile(filePath, []byte("{not-a-valid-json-array}"), 0644); err != nil {
			t.Fatalf("failed to write corrupted file: %v", err)
		}

		_, err := repository.NewFileRepository(filePath)
		if err == nil {
			t.Fatal("expected error on corrupted JSON, got nil")
		}
		if !errors.Is(err, model.ErrStorageFailure) {
			t.Errorf("expected error wrapping ErrStorageFailure, got %v", err)
		}
	})

	t.Run("returns ErrStorageFailure when directory creation fails", func(t *testing.T) {
		tempDir := t.TempDir()
		// Create a file where a directory should be
		conflictFile := filepath.Join(tempDir, "conflict")
		if err := os.WriteFile(conflictFile, []byte("blocker"), 0644); err != nil {
			t.Fatalf("failed to create conflict file: %v", err)
		}

		invalidPath := filepath.Join(conflictFile, "tasks.json")
		_, err := repository.NewFileRepository(invalidPath)
		if err == nil {
			t.Fatal("expected error creating directory inside regular file, got nil")
		}
		if !errors.Is(err, model.ErrStorageFailure) {
			t.Errorf("expected error wrapping ErrStorageFailure, got %v", err)
		}
	})
}

func TestFileRepository_CRUD_And_Persistence(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "persistence_test.json")

	repo, err := repository.NewFileRepository(filePath)
	if err != nil {
		t.Fatalf("failed to initialize FileRepository: %v", err)
	}

	task1 := &model.Task{
		ID:          "file-task-1",
		Title:       "File Persistence Task 1",
		Description: "Testing file store persistence",
		Status:      model.StatusTodo,
		Priority:    model.PriorityHigh,
		CreatedAt:   time.Now().UTC().Add(-2 * time.Hour),
		UpdatedAt:   time.Now().UTC().Add(-2 * time.Hour),
	}
	task2 := &model.Task{
		ID:          "file-task-2",
		Title:       "File Persistence Task 2",
		Description: "Searchable keyword query",
		Status:      model.StatusInProgress,
		Priority:    model.PriorityMedium,
		CreatedAt:   time.Now().UTC().Add(-1 * time.Hour),
		UpdatedAt:   time.Now().UTC().Add(-1 * time.Hour),
	}

	// 1. Create
	t.Run("Create tasks", func(t *testing.T) {
		if err := repo.Create(ctx, task1); err != nil {
			t.Fatalf("failed to create task1: %v", err)
		}
		if err := repo.Create(ctx, task2); err != nil {
			t.Fatalf("failed to create task2: %v", err)
		}

		// Conflict check
		err := repo.Create(ctx, task1)
		if !errors.Is(err, model.ErrTaskConflict) {
			t.Errorf("expected ErrTaskConflict, got %v", err)
		}

		// Nil task check
		err = repo.Create(ctx, nil)
		if !errors.Is(err, model.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}

		// Canceled context check
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		err = repo.Create(canceledCtx, &model.Task{ID: "canceled"})
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	// 2. GetByID
	t.Run("GetByID tasks", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "file-task-1")
		if err != nil {
			t.Fatalf("failed to get task1: %v", err)
		}
		if got.Title != task1.Title {
			t.Errorf("got title %q, want %q", got.Title, task1.Title)
		}

		// Non-existent
		_, err = repo.GetByID(ctx, "non-existent")
		if !errors.Is(err, model.ErrTaskNotFound) {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}

		// Canceled context
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err = repo.GetByID(canceledCtx, "file-task-1")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	// 3. List
	t.Run("List with filtering", func(t *testing.T) {
		// All
		all, err := repo.List(ctx, model.TaskFilter{})
		if err != nil {
			t.Fatalf("failed to list all: %v", err)
		}
		if len(all) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(all))
		}
		// Descending order by CreatedAt: task2 is newer (-1h) than task1 (-2h)
		if all[0].ID != "file-task-2" || all[1].ID != "file-task-1" {
			t.Errorf("expected order [task2, task1], got [%s, %s]", all[0].ID, all[1].ID)
		}

		// Filter status
		inProg, err := repo.List(ctx, model.TaskFilter{Status: model.StatusInProgress})
		if err != nil {
			t.Fatalf("failed to filter status: %v", err)
		}
		if len(inProg) != 1 || inProg[0].ID != "file-task-2" {
			t.Errorf("expected 1 in-progress task, got %v", inProg)
		}

		// Filter search
		searchRes, err := repo.List(ctx, model.TaskFilter{Search: "keyword"})
		if err != nil {
			t.Fatalf("failed to search: %v", err)
		}
		if len(searchRes) != 1 || searchRes[0].ID != "file-task-2" {
			t.Errorf("expected 1 search result, got %v", searchRes)
		}

		// Canceled context
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		_, err = repo.List(canceledCtx, model.TaskFilter{})
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	// 4. Update
	t.Run("Update task", func(t *testing.T) {
		updated := &model.Task{
			ID:          "file-task-1",
			Title:       "Updated Title In File",
			Description: "Updated Description",
			Status:      model.StatusDone,
			Priority:    model.PriorityLow,
			CreatedAt:   task1.CreatedAt,
			UpdatedAt:   time.Now().UTC(),
		}

		if err := repo.Update(ctx, updated); err != nil {
			t.Fatalf("failed to update task1: %v", err)
		}

		got, err := repo.GetByID(ctx, "file-task-1")
		if err != nil {
			t.Fatalf("failed to get updated task: %v", err)
		}
		if got.Title != "Updated Title In File" || got.Status != model.StatusDone {
			t.Errorf("updated task has unexpected values: %+v", got)
		}

		// Nil task
		err = repo.Update(ctx, nil)
		if !errors.Is(err, model.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}

		// Non-existent task
		err = repo.Update(ctx, &model.Task{ID: "non-existent"})
		if !errors.Is(err, model.ErrTaskNotFound) {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}

		// Canceled context
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		err = repo.Update(canceledCtx, updated)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	// 5. File Reload & Persistence Verification
	t.Run("Persistence across new repository instance", func(t *testing.T) {
		// Instantiate a fresh repository pointing to the same file
		reloadedRepo, err := repository.NewFileRepository(filePath)
		if err != nil {
			t.Fatalf("failed to reload repository from file: %v", err)
		}

		// Check task1 updated values are preserved
		t1, err := reloadedRepo.GetByID(ctx, "file-task-1")
		if err != nil {
			t.Fatalf("failed to get reloaded task1: %v", err)
		}
		if t1.Title != "Updated Title In File" || t1.Status != model.StatusDone {
			t.Errorf("reloaded task1 values do not match: %+v", t1)
		}

		// Check task2 is present
		t2, err := reloadedRepo.GetByID(ctx, "file-task-2")
		if err != nil {
			t.Fatalf("failed to get reloaded task2: %v", err)
		}
		if t2.Title != task2.Title {
			t.Errorf("reloaded task2 title mismatch: got %q, want %q", t2.Title, task2.Title)
		}

		// Verify no leftover .tmp file
		tmpPath := filePath + ".tmp"
		if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
			t.Errorf("temporary file %q was not cleaned up", tmpPath)
		}
	})

	// 6. Delete
	t.Run("Delete task", func(t *testing.T) {
		if err := repo.Delete(ctx, "file-task-2"); err != nil {
			t.Fatalf("failed to delete task2: %v", err)
		}

		_, err := repo.GetByID(ctx, "file-task-2")
		if !errors.Is(err, model.ErrTaskNotFound) {
			t.Errorf("expected ErrTaskNotFound after deletion, got %v", err)
		}

		// Non-existent delete
		err = repo.Delete(ctx, "file-task-2")
		if !errors.Is(err, model.ErrTaskNotFound) {
			t.Errorf("expected ErrTaskNotFound deleting already deleted task, got %v", err)
		}

		// Canceled context
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		err = repo.Delete(canceledCtx, "file-task-1")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}

		// Verify deletion persists in a reloaded repo
		freshRepo, err := repository.NewFileRepository(filePath)
		if err != nil {
			t.Fatalf("failed to reload repo after delete: %v", err)
		}
		tasks, err := freshRepo.List(ctx, model.TaskFilter{})
		if err != nil {
			t.Fatalf("failed to list reloaded repo: %v", err)
		}
		if len(tasks) != 1 || tasks[0].ID != "file-task-1" {
			t.Errorf("expected exactly 1 task remaining after delete, got %d", len(tasks))
		}
	})
}

func TestFileRepository_ConcurrencyRace(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "concurrent_file_tasks.json")

	repo, err := repository.NewFileRepository(filePath)
	if err != nil {
		t.Fatalf("failed to initialize file repo: %v", err)
	}

	ctx := context.Background()
	const numGoroutines = 20
	const opsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		goroutineID := g
		go func() {
			defer wg.Done()

			for op := 0; op < opsPerGoroutine; op++ {
				id := fmt.Sprintf("file-concurrent-%d-%d", goroutineID, op)
				task := &model.Task{
					ID:          id,
					Title:       fmt.Sprintf("Title %s", id),
					Description: "Concurrent file task",
					Status:      model.StatusTodo,
					Priority:    model.PriorityMedium,
					CreatedAt:   time.Now().UTC(),
					UpdatedAt:   time.Now().UTC(),
				}

				_ = repo.Create(ctx, task)
				_, _ = repo.GetByID(ctx, id)
				_, _ = repo.List(ctx, model.TaskFilter{Status: model.StatusTodo})
			}
		}()
	}

	wg.Wait()

	// Verify all tasks can be read from a reloaded repo
	reloaded, err := repository.NewFileRepository(filePath)
	if err != nil {
		t.Fatalf("failed to reload repo after concurrent writes: %v", err)
	}

	all, err := reloaded.List(ctx, model.TaskFilter{})
	if err != nil {
		t.Fatalf("failed to list reloaded repo: %v", err)
	}

	if len(all) == 0 {
		t.Error("expected written tasks in reloaded repo, got 0")
	}
}
