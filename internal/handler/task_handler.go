package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/service"
)

// CreateTaskRequest defines the incoming payload for creating a task.
type CreateTaskRequest struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Priority    model.Priority `json:"priority"`
}

// UpdateTaskRequest defines the incoming payload for updating a task.
type UpdateTaskRequest struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      model.Status   `json:"status"`
	Priority    model.Priority `json:"priority"`
}

// TaskHandler manages HTTP endpoints for Task management.
type TaskHandler struct {
	service service.TaskService
	logger  *slog.Logger
}

// NewTaskHandler constructs a TaskHandler with the provided service and logger.
func NewTaskHandler(svc service.TaskService, logger *slog.Logger) *TaskHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &TaskHandler{
		service: svc,
		logger:  logger,
	}
}

// RegisterRoutes registers endpoints using Go 1.22+ pattern matching on ServeMux.
func (h *TaskHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tasks", h.CreateTask)
	mux.HandleFunc("GET /api/v1/tasks", h.ListTasks)
	mux.HandleFunc("GET /api/v1/tasks/{id}", h.GetTask)
	mux.HandleFunc("PUT /api/v1/tasks/{id}", h.UpdateTask)
	mux.HandleFunc("DELETE /api/v1/tasks/{id}", h.DeleteTask)
	mux.HandleFunc("GET /health", h.HealthCheck)
}

// CreateTask handles POST /api/v1/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := ReadJSON(w, r, &req); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.service.CreateTask(r.Context(), req.Title, req.Description, req.Priority)
	if err != nil {
		if errors.Is(err, model.ErrInvalidInput) {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("Failed to create task", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	JSON(w, http.StatusCreated, task)
}

// ListTasks handles GET /api/v1/tasks with optional query filtering
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := model.TaskFilter{
		Status:   model.Status(query.Get("status")),
		Priority: model.Priority(query.Get("priority")),
		Search:   query.Get("search"),
	}

	tasks, err := h.service.ListTasks(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list tasks", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	JSON(w, http.StatusOK, tasks)
}

// GetTask handles GET /api/v1/tasks/{id} using Go 1.22+ r.PathValue
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, err := h.service.GetTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrTaskNotFound) {
			Error(w, http.StatusNotFound, "task not found")
			return
		}
		if errors.Is(err, model.ErrInvalidInput) {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("Failed to get task", "id", id, "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	JSON(w, http.StatusOK, task)
}

// UpdateTask handles PUT /api/v1/tasks/{id}
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req UpdateTaskRequest
	if err := ReadJSON(w, r, &req); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.service.UpdateTask(r.Context(), id, req.Title, req.Description, req.Status, req.Priority)
	if err != nil {
		if errors.Is(err, model.ErrTaskNotFound) {
			Error(w, http.StatusNotFound, "task not found")
			return
		}
		if errors.Is(err, model.ErrInvalidInput) {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("Failed to update task", "id", id, "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	JSON(w, http.StatusOK, task)
}

// DeleteTask handles DELETE /api/v1/tasks/{id}
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.service.DeleteTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrTaskNotFound) {
			Error(w, http.StatusNotFound, "task not found")
			return
		}
		if errors.Is(err, model.ErrInvalidInput) {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("Failed to delete task", "id", id, "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HealthCheck handles GET /health
func (h *TaskHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
