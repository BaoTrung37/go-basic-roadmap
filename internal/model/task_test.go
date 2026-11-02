package model_test

import (
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"go-basic-roadmap/internal/model"
)

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status model.Status
		want   bool
	}{
		{name: "status TODO is valid", status: model.StatusTodo, want: true},
		{name: "status IN_PROGRESS is valid", status: model.StatusInProgress, want: true},
		{name: "status DONE is valid", status: model.StatusDone, want: true},
		{name: "empty status is invalid", status: model.Status(""), want: false},
		{name: "lowercase todo is invalid", status: model.Status("todo"), want: false},
		{name: "random string status is invalid", status: model.Status("PENDING"), want: false},
		{name: "whitespace status is invalid", status: model.Status(" "), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := model.IsValidStatus(tc.status)
			if got != tc.want {
				t.Errorf("IsValidStatus(%q) = %v, want %v", tc.status, got, tc.want)
			}
		})
	}
}

func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority model.Priority
		want     bool
	}{
		{name: "priority LOW is valid", priority: model.PriorityLow, want: true},
		{name: "priority MEDIUM is valid", priority: model.PriorityMedium, want: true},
		{name: "priority HIGH is valid", priority: model.PriorityHigh, want: true},
		{name: "empty priority is invalid", priority: model.Priority(""), want: false},
		{name: "lowercase low is invalid", priority: model.Priority("low"), want: false},
		{name: "arbitrary string is invalid", priority: model.Priority("URGENT"), want: false},
		{name: "whitespace priority is invalid", priority: model.Priority(" "), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := model.IsValidPriority(tc.priority)
			if got != tc.want {
				t.Errorf("IsValidPriority(%q) = %v, want %v", tc.priority, got, tc.want)
			}
		})
	}
}

func TestValidateCreateTask(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		priority   model.Priority
		wantErr    bool
		expectedIs error
	}{
		{
			name:     "valid title and empty priority",
			title:    "Write comprehensive unit tests",
			priority: "",
			wantErr:  false,
		},
		{
			name:     "valid title with low priority",
			title:    "Clean desk",
			priority: model.PriorityLow,
			wantErr:  false,
		},
		{
			name:     "valid title with medium priority",
			title:    "Refactor handler",
			priority: model.PriorityMedium,
			wantErr:  false,
		},
		{
			name:     "valid title with high priority",
			title:    "Fix critical race condition",
			priority: model.PriorityHigh,
			wantErr:  false,
		},
		{
			name:     "boundary: title exactly 100 characters",
			title:    strings.Repeat("a", 100),
			priority: model.PriorityMedium,
			wantErr:  false,
		},
		{
			name:       "error: empty title",
			title:      "",
			priority:   model.PriorityMedium,
			wantErr:    true,
			expectedIs: model.ErrTitleRequired,
		},
		{
			name:       "error: whitespace-only title",
			title:      "   \t\n   ",
			priority:   model.PriorityMedium,
			wantErr:    true,
			expectedIs: model.ErrTitleRequired,
		},
		{
			name:       "error: title exceeds 100 characters",
			title:      strings.Repeat("a", 101),
			priority:   model.PriorityLow,
			wantErr:    true,
			expectedIs: model.ErrTitleTooLong,
		},
		{
			name:       "error: invalid priority string",
			title:      "Implement feature",
			priority:   model.Priority("SUPER_HIGH"),
			wantErr:    true,
			expectedIs: model.ErrInvalidPriority,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := model.ValidateCreateTask(tc.title, tc.priority)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ValidateCreateTask(%q, %q) expected error, got nil", tc.title, tc.priority)
				}
				if !errors.Is(err, model.ErrInvalidInput) {
					t.Errorf("expected error wrapping ErrInvalidInput, got %v", err)
				}
				if tc.expectedIs != nil && !errors.Is(err, tc.expectedIs) {
					t.Errorf("expected error wrapping %v, got %v", tc.expectedIs, err)
				}
			} else {
				if err != nil {
					t.Fatalf("ValidateCreateTask(%q, %q) unexpected error: %v", tc.title, tc.priority, err)
				}
			}
		})
	}
}

func TestValidateUpdateTask(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		status     model.Status
		priority   model.Priority
		wantErr    bool
		expectedIs error
	}{
		{
			name:     "valid update to IN_PROGRESS and HIGH",
			title:    "Deploy to staging",
			status:   model.StatusInProgress,
			priority: model.PriorityHigh,
			wantErr:  false,
		},
		{
			name:     "valid update to DONE and LOW",
			title:    "Archived task",
			status:   model.StatusDone,
			priority: model.PriorityLow,
			wantErr:  false,
		},
		{
			name:     "boundary: title exactly 100 characters",
			title:    strings.Repeat("z", 100),
			status:   model.StatusTodo,
			priority: model.PriorityMedium,
			wantErr:  false,
		},
		{
			name:       "error: empty title",
			title:      "",
			status:     model.StatusTodo,
			priority:   model.PriorityMedium,
			wantErr:    true,
			expectedIs: model.ErrTitleRequired,
		},
		{
			name:       "error: whitespace title",
			title:      "     ",
			status:     model.StatusTodo,
			priority:   model.PriorityMedium,
			wantErr:    true,
			expectedIs: model.ErrTitleRequired,
		},
		{
			name:       "error: title exceeds 100 characters",
			title:      strings.Repeat("z", 101),
			status:     model.StatusTodo,
			priority:   model.PriorityMedium,
			wantErr:    true,
			expectedIs: model.ErrTitleTooLong,
		},
		{
			name:       "error: invalid status",
			title:      "Valid Title",
			status:     model.Status("BLOCKED"),
			priority:   model.PriorityMedium,
			wantErr:    true,
			expectedIs: model.ErrInvalidStatus,
		},
		{
			name:       "error: empty status",
			title:      "Valid Title",
			status:     model.Status(""),
			priority:   model.PriorityMedium,
			wantErr:    true,
			expectedIs: model.ErrInvalidStatus,
		},
		{
			name:       "error: invalid priority",
			title:      "Valid Title",
			status:     model.StatusInProgress,
			priority:   model.Priority("CRITICAL"),
			wantErr:    true,
			expectedIs: model.ErrInvalidPriority,
		},
		{
			name:       "error: empty priority",
			title:      "Valid Title",
			status:     model.StatusInProgress,
			priority:   model.Priority(""),
			wantErr:    true,
			expectedIs: model.ErrInvalidPriority,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := model.ValidateUpdateTask(tc.title, tc.status, tc.priority)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ValidateUpdateTask expected error, got nil")
				}
				if !errors.Is(err, model.ErrInvalidInput) {
					t.Errorf("expected error wrapping ErrInvalidInput, got %v", err)
				}
				if tc.expectedIs != nil && !errors.Is(err, tc.expectedIs) {
					t.Errorf("expected error wrapping %v, got %v", tc.expectedIs, err)
				}
			} else {
				if err != nil {
					t.Fatalf("ValidateUpdateTask unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTask_Clone(t *testing.T) {
	t.Run("nil receiver returns nil", func(t *testing.T) {
		var task *model.Task
		cloned := task.Clone()
		if cloned != nil {
			t.Errorf("expected nil clone for nil receiver, got %v", cloned)
		}
	})

	t.Run("deep copy produces independent struct", func(t *testing.T) {
		now := time.Now().UTC()
		original := &model.Task{
			ID:          "task-123",
			Title:       "Original Title",
			Description: "Original Description",
			Status:      model.StatusTodo,
			Priority:    model.PriorityMedium,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		cloned := original.Clone()

		if cloned == nil {
			t.Fatal("expected non-nil clone")
		}

		// Verify pointer identity is different
		if cloned == original {
			t.Error("cloned task shares the same memory address as original")
		}

		// Verify fields are identical
		if cloned.ID != original.ID ||
			cloned.Title != original.Title ||
			cloned.Description != original.Description ||
			cloned.Status != original.Status ||
			cloned.Priority != original.Priority ||
			!cloned.CreatedAt.Equal(original.CreatedAt) ||
			!cloned.UpdatedAt.Equal(original.UpdatedAt) {
			t.Errorf("cloned fields do not match original: got %+v, want %+v", cloned, original)
		}

		// Verify mutation does not affect original
		cloned.Title = "Mutated Title"
		cloned.Status = model.StatusDone

		if original.Title != "Original Title" {
			t.Errorf("mutating clone modified original title: got %q, want 'Original Title'", original.Title)
		}
		if original.Status != model.StatusTodo {
			t.Errorf("mutating clone modified original status: got %q, want 'TODO'", original.Status)
		}
	})
}

func TestGenerateID(t *testing.T) {
	// RFC 4122 v4 UUID format regex: 8-4-4-4-12 hex with version 4 and variant 8,9,a,b
	uuidV4Regex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	t.Run("conforms to RFC 4122 v4 UUID format", func(t *testing.T) {
		id, err := model.GenerateID()
		if err != nil {
			t.Fatalf("GenerateID failed: %v", err)
		}

		if len(id) != 36 {
			t.Errorf("expected UUID length 36, got %d for %q", len(id), id)
		}

		if !uuidV4Regex.MatchString(id) {
			t.Errorf("generated ID %q does not match RFC 4122 v4 UUID pattern", id)
		}
	})

	t.Run("generates unique IDs with no collisions", func(t *testing.T) {
		const count = 1000
		seen := make(map[string]struct{}, count)

		for i := 0; i < count; i++ {
			id, err := model.GenerateID()
			if err != nil {
				t.Fatalf("GenerateID failed at iteration %d: %v", i, err)
			}
			if _, exists := seen[id]; exists {
				t.Fatalf("collision detected for ID %q at iteration %d", id, i)
			}
			seen[id] = struct{}{}
		}

		if len(seen) != count {
			t.Errorf("expected %d unique IDs, got %d", count, len(seen))
		}
	})
}
