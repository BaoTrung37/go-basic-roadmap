package model

import "errors"

// Domain sentinel errors for TaskManager.
var (
	// ErrTaskNotFound is returned when a task with the requested ID does not exist.
	ErrTaskNotFound = errors.New("task not found")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrTaskConflict is returned when attempting to create a task that already exists.
	ErrTaskConflict = errors.New("task conflict: task with given ID already exists")

	// ErrStorageFailure is returned when an underlying I/O or storage operation fails.
	ErrStorageFailure = errors.New("storage failure")

	// Specific input validation errors
	ErrTitleRequired   = errors.New("title cannot be empty")
	ErrTitleTooLong    = errors.New("title must not exceed 100 characters")
	ErrInvalidStatus   = errors.New("invalid status: must be TODO, IN_PROGRESS, or DONE")
	ErrInvalidPriority = errors.New("invalid priority: must be LOW, MEDIUM, or HIGH")
)
