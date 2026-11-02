package model

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

// Status represents the current progression stage of a Task.
type Status string

const (
	StatusTodo       Status = "TODO"
	StatusInProgress Status = "IN_PROGRESS"
	StatusDone       Status = "DONE"
)

// Priority represents the execution urgency of a Task.
type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityMedium Priority = "MEDIUM"
	PriorityHigh   Priority = "HIGH"
)

// Task is the core domain entity representing a work item.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Clone creates a deep copy of a Task to avoid concurrency pointer sharing.
func (t *Task) Clone() *Task {
	if t == nil {
		return nil
	}
	copied := *t
	return &copied
}

// TaskFilter defines criteria for querying and filtering tasks.
type TaskFilter struct {
	Status   Status
	Priority Priority
	Search   string
}

// Event types dispatched across the worker pool.
const (
	EventTypeCreated       = "CREATED"
	EventTypeUpdated       = "UPDATED"
	EventTypeDeleted       = "DELETED"
	EventTypeStatusChanged = "STATUS_CHANGED"
)

// TaskEvent captures a lifecycle event emitted when a Task is modified.
type TaskEvent struct {
	Type      string    `json:"type"` // CREATED, UPDATED, DELETED, STATUS_CHANGED
	Task      Task      `json:"task"`
	Timestamp time.Time `json:"timestamp"`
}

// IsValidStatus checks if a given status belongs to allowed enum values.
func IsValidStatus(s Status) bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// IsValidPriority checks if a given priority belongs to allowed enum values.
func IsValidPriority(p Priority) bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

// ValidateCreateTask validates mandatory fields when creating a task.
func ValidateCreateTask(title string, priority Priority) error {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return fmt.Errorf("%w: %w", ErrInvalidInput, ErrTitleRequired)
	}
	if len(trimmed) > 100 {
		return fmt.Errorf("%w: %w", ErrInvalidInput, ErrTitleTooLong)
	}
	if priority != "" && !IsValidPriority(priority) {
		return fmt.Errorf("%w: %w", ErrInvalidInput, ErrInvalidPriority)
	}
	return nil
}

// ValidateUpdateTask validates fields when updating an existing task.
func ValidateUpdateTask(title string, status Status, priority Priority) error {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return fmt.Errorf("%w: %w", ErrInvalidInput, ErrTitleRequired)
	}
	if len(trimmed) > 100 {
		return fmt.Errorf("%w: %w", ErrInvalidInput, ErrTitleTooLong)
	}
	if !IsValidStatus(status) {
		return fmt.Errorf("%w: %w", ErrInvalidInput, ErrInvalidStatus)
	}
	if !IsValidPriority(priority) {
		return fmt.Errorf("%w: %w", ErrInvalidInput, ErrInvalidPriority)
	}
	return nil
}

// GenerateID produces an RFC 4122 v4 UUID string using crypto/rand.
func GenerateID() (string, error) {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		return "", err
	}

	// Set version 4 (0100) in the upper 4 bits of byte 6
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// Set variant 10 in the upper 2 bits of byte 8
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}
