package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-basic-roadmap/internal/handler"
	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/service"
)

// discardLogger creates a logger that discards all log messages for clean tests.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// mockTaskService implements service.TaskService for isolated handler tests.
type mockTaskService struct {
	createTaskFunc func(ctx context.Context, title, description string, priority model.Priority) (*model.Task, error)
	getTaskFunc    func(ctx context.Context, id string) (*model.Task, error)
	listTasksFunc  func(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error)
	updateTaskFunc func(ctx context.Context, id, title, description string, status model.Status, priority model.Priority) (*model.Task, error)
	deleteTaskFunc func(ctx context.Context, id string) error
}

func (m *mockTaskService) CreateTask(ctx context.Context, title, description string, priority model.Priority) (*model.Task, error) {
	if m.createTaskFunc != nil {
		return m.createTaskFunc(ctx, title, description, priority)
	}
	return nil, nil
}

func (m *mockTaskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	if m.getTaskFunc != nil {
		return m.getTaskFunc(ctx, id)
	}
	return nil, model.ErrTaskNotFound
}

func (m *mockTaskService) ListTasks(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
	if m.listTasksFunc != nil {
		return m.listTasksFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockTaskService) UpdateTask(ctx context.Context, id, title, description string, status model.Status, priority model.Priority) (*model.Task, error) {
	if m.updateTaskFunc != nil {
		return m.updateTaskFunc(ctx, id, title, description, status, priority)
	}
	return nil, model.ErrTaskNotFound
}

func (m *mockTaskService) DeleteTask(ctx context.Context, id string) error {
	if m.deleteTaskFunc != nil {
		return m.deleteTaskFunc(ctx, id)
	}
	return nil
}

var _ service.TaskService = (*mockTaskService)(nil)

func TestNewTaskHandler(t *testing.T) {
	t.Run("constructs with default logger when nil logger is provided", func(t *testing.T) {
		h := handler.NewTaskHandler(&mockTaskService{}, nil)
		if h == nil {
			t.Fatal("expected non-nil TaskHandler")
		}
	})
}

func TestTaskHandler_CreateTask(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockFunc       func(ctx context.Context, title, description string, priority model.Priority) (*model.Task, error)
		expectedStatus int
		expectedSub    string
	}{
		{
			name:        "201 Created on valid request",
			requestBody: `{"title":"Learn Go","description":"Study standard library","priority":"HIGH"}`,
			mockFunc: func(ctx context.Context, title, description string, priority model.Priority) (*model.Task, error) {
				return &model.Task{
					ID:          "task-new-1",
					Title:       title,
					Description: description,
					Status:      model.StatusTodo,
					Priority:    priority,
					CreatedAt:   time.Now().UTC(),
					UpdatedAt:   time.Now().UTC(),
				}, nil
			},
			expectedStatus: http.StatusCreated,
			expectedSub:    `"id":"task-new-1"`,
		},
		{
			name:           "400 Bad Request on empty request body",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "request body must not be empty",
		},
		{
			name:           "400 Bad Request on malformed JSON",
			requestBody:    `{"title": "Unclosed string}`,
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "malformed JSON",
		},
		{
			name:           "400 Bad Request on unknown fields",
			requestBody:    `{"title":"Task","extra_unknown_field":"bad"}`,
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "unknown field",
		},
		{
			name:        "400 Bad Request on service validation failure",
			requestBody: `{"title":"   ","priority":"LOW"}`,
			mockFunc: func(ctx context.Context, title, description string, priority model.Priority) (*model.Task, error) {
				return nil, fmt.Errorf("%w: %w", model.ErrInvalidInput, model.ErrTitleRequired)
			},
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "title cannot be empty",
		},
		{
			name:        "500 Internal Server Error on unexpected service failure",
			requestBody: `{"title":"Valid Title"}`,
			mockFunc: func(ctx context.Context, title, description string, priority model.Priority) (*model.Task, error) {
				return nil, errors.New("database connection down")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedSub:    "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockTaskService{
				createTaskFunc: tc.mockFunc,
			}
			h := handler.NewTaskHandler(svc, discardLogger())

			req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.CreateTask(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("status code = %d, want %d; body = %s", rec.Code, tc.expectedStatus, rec.Body.String())
			}
			if tc.expectedSub != "" && !strings.Contains(rec.Body.String(), tc.expectedSub) {
				t.Errorf("response body %q does not contain expected substring %q", rec.Body.String(), tc.expectedSub)
			}
		})
	}
}

func TestTaskHandler_ListTasks(t *testing.T) {
	t.Run("200 OK returns task list and forwards query parameters", func(t *testing.T) {
		var receivedFilter model.TaskFilter
		mockTasks := []*model.Task{
			{ID: "task-1", Title: "Task 1", Status: model.StatusTodo, Priority: model.PriorityHigh},
			{ID: "task-2", Title: "Task 2", Status: model.StatusTodo, Priority: model.PriorityMedium},
		}

		svc := &mockTaskService{
			listTasksFunc: func(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
				receivedFilter = filter
				return mockTasks, nil
			},
		}

		h := handler.NewTaskHandler(svc, discardLogger())

		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?status=TODO&priority=HIGH&search=Task", nil)
		rec := httptest.NewRecorder()

		h.ListTasks(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
		}

		if receivedFilter.Status != model.StatusTodo {
			t.Errorf("expected filter Status 'TODO', got %q", receivedFilter.Status)
		}
		if receivedFilter.Priority != model.PriorityHigh {
			t.Errorf("expected filter Priority 'HIGH', got %q", receivedFilter.Priority)
		}
		if receivedFilter.Search != "Task" {
			t.Errorf("expected filter Search 'Task', got %q", receivedFilter.Search)
		}

		var responseData []*model.Task
		if err := json.Unmarshal(rec.Body.Bytes(), &responseData); err != nil {
			t.Fatalf("failed to parse response JSON: %v", err)
		}
		if len(responseData) != 2 {
			t.Errorf("expected 2 tasks in response, got %d", len(responseData))
		}
	})

	t.Run("500 Internal Server Error when service fails", func(t *testing.T) {
		svc := &mockTaskService{
			listTasksFunc: func(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
				return nil, errors.New("read storage error")
			},
		}

		h := handler.NewTaskHandler(svc, discardLogger())
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
		rec := httptest.NewRecorder()

		h.ListTasks(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status code = %d, want 500", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "internal server error") {
			t.Errorf("expected 'internal server error' in body, got %s", rec.Body.String())
		}
	})
}

func TestTaskHandler_GetTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		mockFunc       func(ctx context.Context, id string) (*model.Task, error)
		expectedStatus int
		expectedSub    string
	}{
		{
			name:   "200 OK for existing task",
			taskID: "task-abc",
			mockFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return &model.Task{
					ID:    id,
					Title: "Retrieved Task",
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedSub:    `"id":"task-abc"`,
		},
		{
			name:   "404 Not Found for missing task",
			taskID: "non-existent",
			mockFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return nil, model.ErrTaskNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedSub:    "task not found",
		},
		{
			name:   "400 Bad Request on empty or invalid ID",
			taskID: "",
			mockFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return nil, fmt.Errorf("%w: ID is invalid", model.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "invalid",
		},
		{
			name:   "500 Internal Server Error on unexpected service failure",
			taskID: "task-crash",
			mockFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return nil, errors.New("disk failure")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedSub:    "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockTaskService{
				getTaskFunc: tc.mockFunc,
			}
			h := handler.NewTaskHandler(svc, discardLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+tc.taskID, nil)
			req.SetPathValue("id", tc.taskID)
			rec := httptest.NewRecorder()

			h.GetTask(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("status code = %d, want %d; body: %s", rec.Code, tc.expectedStatus, rec.Body.String())
			}
			if tc.expectedSub != "" && !strings.Contains(rec.Body.String(), tc.expectedSub) {
				t.Errorf("body %q does not contain %q", rec.Body.String(), tc.expectedSub)
			}
		})
	}
}

func TestTaskHandler_UpdateTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		body           string
		mockFunc       func(ctx context.Context, id, title, description string, status model.Status, priority model.Priority) (*model.Task, error)
		expectedStatus int
		expectedSub    string
	}{
		{
			name:   "200 OK on successful update",
			taskID: "task-update-1",
			body:   `{"title":"Updated Title","description":"Updated Desc","status":"IN_PROGRESS","priority":"HIGH"}`,
			mockFunc: func(ctx context.Context, id, title, description string, status model.Status, priority model.Priority) (*model.Task, error) {
				return &model.Task{
					ID:          id,
					Title:       title,
					Description: description,
					Status:      status,
					Priority:    priority,
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedSub:    `"title":"Updated Title"`,
		},
		{
			name:           "400 Bad Request on invalid JSON body",
			taskID:         "task-update-1",
			body:           `{bad-json`,
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "malformed JSON",
		},
		{
			name:   "404 Not Found when task does not exist",
			taskID: "missing-task",
			body:   `{"title":"Updated Title","status":"DONE","priority":"LOW"}`,
			mockFunc: func(ctx context.Context, id, title, description string, status model.Status, priority model.Priority) (*model.Task, error) {
				return nil, model.ErrTaskNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedSub:    "task not found",
		},
		{
			name:   "400 Bad Request on validation failure",
			taskID: "task-1",
			body:   `{"title":"","status":"DONE","priority":"LOW"}`,
			mockFunc: func(ctx context.Context, id, title, description string, status model.Status, priority model.Priority) (*model.Task, error) {
				return nil, fmt.Errorf("%w: %w", model.ErrInvalidInput, model.ErrTitleRequired)
			},
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "title cannot be empty",
		},
		{
			name:   "500 Internal Server Error on unexpected failure",
			taskID: "task-1",
			body:   `{"title":"Title","status":"DONE","priority":"LOW"}`,
			mockFunc: func(ctx context.Context, id, title, description string, status model.Status, priority model.Priority) (*model.Task, error) {
				return nil, errors.New("storage timeout")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedSub:    "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockTaskService{
				updateTaskFunc: tc.mockFunc,
			}
			h := handler.NewTaskHandler(svc, discardLogger())

			req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/"+tc.taskID, bytes.NewBufferString(tc.body))
			req.SetPathValue("id", tc.taskID)
			rec := httptest.NewRecorder()

			h.UpdateTask(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("status code = %d, want %d; body = %s", rec.Code, tc.expectedStatus, rec.Body.String())
			}
			if tc.expectedSub != "" && !strings.Contains(rec.Body.String(), tc.expectedSub) {
				t.Errorf("body %q does not contain %q", rec.Body.String(), tc.expectedSub)
			}
		})
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		mockFunc       func(ctx context.Context, id string) error
		expectedStatus int
		expectedSub    string
	}{
		{
			name:   "204 No Content on successful deletion",
			taskID: "task-delete-me",
			mockFunc: func(ctx context.Context, id string) error {
				return nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "404 Not Found on missing task",
			taskID: "task-missing",
			mockFunc: func(ctx context.Context, id string) error {
				return model.ErrTaskNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedSub:    "task not found",
		},
		{
			name:   "400 Bad Request on invalid input ID",
			taskID: "bad-id",
			mockFunc: func(ctx context.Context, id string) error {
				return fmt.Errorf("%w: invalid ID format", model.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "invalid",
		},
		{
			name:   "500 Internal Server Error on unexpected failure",
			taskID: "task-crash",
			mockFunc: func(ctx context.Context, id string) error {
				return errors.New("unexpected error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedSub:    "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockTaskService{
				deleteTaskFunc: tc.mockFunc,
			}
			h := handler.NewTaskHandler(svc, discardLogger())

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+tc.taskID, nil)
			req.SetPathValue("id", tc.taskID)
			rec := httptest.NewRecorder()

			h.DeleteTask(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("status code = %d, want %d; body = %s", rec.Code, tc.expectedStatus, rec.Body.String())
			}
			if tc.expectedSub != "" && !strings.Contains(rec.Body.String(), tc.expectedSub) {
				t.Errorf("body %q does not contain %q", rec.Body.String(), tc.expectedSub)
			}
		})
	}
}

func TestTaskHandler_HealthCheck(t *testing.T) {
	h := handler.NewTaskHandler(&mockTaskService{}, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.HealthCheck(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode health check response: %v", err)
	}

	if payload["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", payload["status"])
	}
	if _, ok := payload["timestamp"]; !ok {
		t.Error("expected 'timestamp' field in health check response")
	}
}

func TestTaskHandler_RegisterRoutes(t *testing.T) {
	// Full router test using ServeMux pattern matching (Go 1.22+)
	svc := &mockTaskService{
		getTaskFunc: func(ctx context.Context, id string) (*model.Task, error) {
			if id == "route-test-123" {
				return &model.Task{ID: id, Title: "Route Match"}, nil
			}
			return nil, model.ErrTaskNotFound
		},
	}

	h := handler.NewTaskHandler(svc, discardLogger())
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	t.Run("router dispatches to GET /api/v1/tasks/{id} and extracts path parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/route-test-123", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK via ServeMux routing, got %d; body: %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), `"id":"route-test-123"`) {
			t.Errorf("unexpected body: %s", rec.Body.String())
		}
	})

	t.Run("router dispatches to GET /health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK on /health, got %d", rec.Code)
		}
	})

	t.Run("router returns 404 on unknown endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 for unknown route, got %d", rec.Code)
		}
	})
}

func TestReadJSON_EdgeCases(t *testing.T) {
	t.Run("rejects body with multiple JSON objects", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"title":"first"}{"title":"second"}`))

		var dst handler.CreateTaskRequest
		err := handler.ReadJSON(rec, req, &dst)
		if err == nil {
			t.Fatal("expected error on multiple JSON objects, got nil")
		}
		if !strings.Contains(err.Error(), "single JSON object") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("rejects payload with invalid type", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"title": 12345}`))

		var dst handler.CreateTaskRequest
		err := handler.ReadJSON(rec, req, &dst)
		if err == nil {
			t.Fatal("expected unmarshal type error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid value for field") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("rejects payload exceeding 1MB", func(t *testing.T) {
		rec := httptest.NewRecorder()
		// Create payload slightly exceeding 1MB
		oversized := strings.Repeat("x", 1048576+100)
		body := fmt.Sprintf(`{"title":"%s"}`, oversized)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

		var dst handler.CreateTaskRequest
		err := handler.ReadJSON(rec, req, &dst)
		if err == nil {
			t.Fatal("expected error on payload > 1MB, got nil")
		}
		if !strings.Contains(err.Error(), "maximum allowed size") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
