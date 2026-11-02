package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"go-basic-roadmap/internal/model"
)

// FileRepository provides atomic JSON file-backed persistence for tasks.
type FileRepository struct {
	mu       sync.RWMutex
	filePath string
	tasks    map[string]*model.Task
}

// NewFileRepository initializes a FileRepository with data loaded from the specified file.
// If the file does not exist, it creates the parent directory and an empty JSON file.
func NewFileRepository(filePath string) (*FileRepository, error) {
	repo := &FileRepository{
		filePath: filePath,
		tasks:    make(map[string]*model.Task),
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("%w: failed to create data directory: %v", model.ErrStorageFailure, err)
	}

	if err := repo.load(); err != nil {
		return nil, err
	}

	return repo, nil
}

// load reads tasks from the JSON file into memory cache.
func (r *FileRepository) load() error {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize with empty array
			return r.save()
		}
		return fmt.Errorf("%w: failed to read file: %v", model.ErrStorageFailure, err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return r.save()
	}

	var taskList []*model.Task
	if err := json.Unmarshal(data, &taskList); err != nil {
		return fmt.Errorf("%w: failed to parse JSON data: %v", model.ErrStorageFailure, err)
	}

	for _, task := range taskList {
		if task != nil && task.ID != "" {
			r.tasks[task.ID] = task
		}
	}

	return nil
}

// save performs an atomic write of all in-memory tasks to the JSON file.
func (r *FileRepository) save() error {
	taskList := make([]*model.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		taskList = append(taskList, task)
	}

	// Sort for deterministic file content
	sort.Slice(taskList, func(i, j int) bool {
		return taskList[i].CreatedAt.Before(taskList[j].CreatedAt)
	})

	encoded, err := json.MarshalIndent(taskList, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: failed to marshal tasks: %v", model.ErrStorageFailure, err)
	}

	tmpFile := r.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, encoded, 0644); err != nil {
		return fmt.Errorf("%w: failed to write temp file: %v", model.ErrStorageFailure, err)
	}

	if err := os.Rename(tmpFile, r.filePath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("%w: failed to atomically replace data file: %v", model.ErrStorageFailure, err)
	}

	return nil
}

// Create persists a new task to memory and disk.
func (r *FileRepository) Create(ctx context.Context, task *model.Task) error {
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
	if err := r.save(); err != nil {
		// Roll back memory state on file failure
		delete(r.tasks, task.ID)
		return err
	}

	return nil
}

// GetByID retrieves a task by ID.
func (r *FileRepository) GetByID(ctx context.Context, id string) (*model.Task, error) {
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

// List returns filtered tasks matching filter criteria.
func (r *FileRepository) List(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
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

// Update updates an existing task in memory and on disk.
func (r *FileRepository) Update(ctx context.Context, task *model.Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if task == nil {
		return model.ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldTask, exists := r.tasks[task.ID]
	if !exists {
		return model.ErrTaskNotFound
	}

	r.tasks[task.ID] = task.Clone()
	if err := r.save(); err != nil {
		// Roll back memory state
		r.tasks[task.ID] = oldTask
		return err
	}

	return nil
}

// Delete removes a task by ID from memory and disk.
func (r *FileRepository) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldTask, exists := r.tasks[id]
	if !exists {
		return model.ErrTaskNotFound
	}

	delete(r.tasks, id)
	if err := r.save(); err != nil {
		// Roll back memory state
		r.tasks[id] = oldTask
		return err
	}

	return nil
}
