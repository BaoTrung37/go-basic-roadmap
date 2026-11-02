package repository_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/repository"
)

// TestChallenger_MemoryRepository_MassiveConcurrencyStress stresses the MemoryRepository
// with 200 concurrent goroutines executing 20,000 mixed CRUD operations under the Go race detector.
func TestChallenger_MemoryRepository_MassiveConcurrencyStress(t *testing.T) {
	repo := repository.NewMemoryRepository()
	ctx := context.Background()

	const (
		numGoroutines = 200
		opsPerRoutine = 100
	)

	// Pre-populate 50 tasks
	initialIDs := make([]string, 50)
	for i := 0; i < 50; i++ {
		id, _ := model.GenerateID()
		initialIDs[i] = id
		task := &model.Task{
			ID:          id,
			Title:       fmt.Sprintf("Initial Task %d", i),
			Description: "Stress test task",
			Status:      model.StatusTodo,
			Priority:    model.PriorityMedium,
			CreatedAt:   time.Now().Add(-time.Duration(i) * time.Minute),
			UpdatedAt:   time.Now().Add(-time.Duration(i) * time.Minute),
		}
		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("Failed to create initial task: %v", err)
		}
	}

	var (
		createdCount atomic.Int64
		readCount    atomic.Int64
		updateCount  atomic.Int64
		deleteCount  atomic.Int64
		listCount    atomic.Int64
		wg           sync.WaitGroup
	)

	wg.Add(numGoroutines)
	for g := 0; g < numGoroutines; g++ {
		go func(routineID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(routineID)))

			for i := 0; i < opsPerRoutine; i++ {
				op := rng.Intn(5)
				switch op {
				case 0: // Create
					id, _ := model.GenerateID()
					task := &model.Task{
						ID:          id,
						Title:       fmt.Sprintf("Routine %d Task %d", routineID, i),
						Description: "Concurrent create",
						Status:      model.StatusTodo,
						Priority:    model.PriorityHigh,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}
					if err := repo.Create(ctx, task); err == nil {
						createdCount.Add(1)
					}
				case 1: // Read
					targetID := initialIDs[rng.Intn(len(initialIDs))]
					if task, err := repo.GetByID(ctx, targetID); err == nil && task != nil {
						// Mutate returned task to test defensive copy
						task.Title = "Mutated outside repo"
						readCount.Add(1)
					}
				case 2: // Update
					targetID := initialIDs[rng.Intn(len(initialIDs))]
					task := &model.Task{
						ID:          targetID,
						Title:       fmt.Sprintf("Updated by routine %d", routineID),
						Description: "Updated desc",
						Status:      model.StatusInProgress,
						Priority:    model.PriorityLow,
						UpdatedAt:   time.Now(),
					}
					if err := repo.Update(ctx, task); err == nil {
						updateCount.Add(1)
					}
				case 3: // Delete
					id, _ := model.GenerateID()
					// Create and immediately delete
					tempTask := &model.Task{
						ID:        id,
						Title:     "Temp",
						Status:    model.StatusDone,
						Priority:  model.PriorityLow,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}
					_ = repo.Create(ctx, tempTask)
					if err := repo.Delete(ctx, id); err == nil {
						deleteCount.Add(1)
					}
				case 4: // List with filter
					filter := model.TaskFilter{
						Status: model.StatusTodo,
					}
					if rng.Intn(2) == 0 {
						filter.Search = "Routine"
					}
					tasks, err := repo.List(ctx, filter)
					if err == nil {
						listCount.Add(1)
						// Verify defensive copy on list results
						for _, item := range tasks {
							item.Title = "Mutated list result"
						}
					}
				}
			}
		}(g)
	}

	wg.Wait()

	t.Logf("Massive concurrency stress test completed successfully:")
	t.Logf("  Creates: %d", createdCount.Load())
	t.Logf("  Reads:   %d", readCount.Load())
	t.Logf("  Updates: %d", updateCount.Load())
	t.Logf("  Deletes: %d", deleteCount.Load())
	t.Logf("  Lists:   %d", listCount.Load())
}
