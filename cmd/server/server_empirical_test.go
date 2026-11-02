package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go-basic-roadmap/internal/handler"
	"go-basic-roadmap/internal/middleware"
	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/repository"
	"go-basic-roadmap/internal/service"
	"go-basic-roadmap/internal/worker"
)

type handlerTransport struct {
	handler http.Handler
}

func (t *handlerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	t.handler.ServeHTTP(rec, req)
	return rec.Result(), nil
}

func buildTestServer(repo repository.TaskRepository, pool worker.WorkerPool, logger *slog.Logger) http.Handler {
	taskService := service.NewTaskService(repo, pool, logger)
	mux := http.NewServeMux()
	taskHandler := handler.NewTaskHandler(taskService, logger)
	taskHandler.RegisterRoutes(mux)

	mux.HandleFunc("GET /api/v1/panic-trigger", func(w http.ResponseWriter, r *http.Request) {
		panic("deliberate test panic for recovery middleware verification")
	})

	var rootHandler http.Handler = mux
	rootHandler = middleware.Recovery(logger)(rootHandler)
	rootHandler = middleware.Logging(logger)(rootHandler)
	rootHandler = middleware.CORS(rootHandler)
	return rootHandler
}

func TestEmpiricalRESTEndpoints(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := repository.NewMemoryRepository()
	pool := worker.NewWorkerPool(4, 100, logger)
	poolCtx, cancelPool := context.WithCancel(context.Background())
	defer cancelPool()
	pool.Start(poolCtx)
	defer pool.Stop()

	rootHandler := buildTestServer(repo, pool, logger)
	client := &http.Client{
		Transport: &handlerTransport{handler: rootHandler},
		Timeout:   5 * time.Second,
	}

	// 1. Health check
	t.Run("GET /health -> 200 OK and healthy status", func(t *testing.T) {
		res, err := client.Get("http://localhost/health")
		if err != nil {
			t.Fatalf("unexpected request error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", res.StatusCode)
		}
		body, _ := io.ReadAll(res.Body)
		var healthData map[string]any
		if err := json.Unmarshal(body, &healthData); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if healthData["status"] != "healthy" {
			t.Fatalf("expected status healthy, got %v", healthData["status"])
		}
		if healthData["timestamp"] == nil {
			t.Fatal("expected timestamp field in health response")
		}
	})

	// 2. CRUD Lifecycle
	var createdID string
	t.Run("POST /api/v1/tasks -> 201 Created on valid input", func(t *testing.T) {
		payload := `{"title":"Empirical Testing Task","description":"Stress test API","priority":"HIGH"}`
		res, err := client.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(payload))
		if err != nil {
			t.Fatalf("POST error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(res.Body)
			t.Fatalf("expected 201 Created, got %d: %s", res.StatusCode, string(body))
		}
		var task model.Task
		if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if task.ID == "" {
			t.Fatal("expected generated task ID")
		}
		if task.Title != "Empirical Testing Task" || task.Status != model.StatusTodo || task.Priority != model.PriorityHigh {
			t.Fatalf("unexpected task fields: %+v", task)
		}
		createdID = task.ID
	})

	t.Run("GET /api/v1/tasks/{id} -> 200 OK and exact match", func(t *testing.T) {
		if createdID == "" {
			t.Fatal("no created task ID available")
		}
		res, err := client.Get("http://localhost/api/v1/tasks/" + createdID)
		if err != nil {
			t.Fatalf("GET error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", res.StatusCode)
		}
		var task model.Task
		if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if task.ID != createdID || task.Title != "Empirical Testing Task" {
			t.Fatalf("task mismatch: %+v", task)
		}
	})

	t.Run("PUT /api/v1/tasks/{id} -> 200 OK and updated fields", func(t *testing.T) {
		if createdID == "" {
			t.Fatal("no created task ID available")
		}
		payload := `{"title":"Updated Empirical Task","description":"Updated description","status":"IN_PROGRESS","priority":"MEDIUM"}`
		req, _ := http.NewRequest(http.MethodPut, "http://localhost/api/v1/tasks/"+createdID, strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("PUT error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(res.Body)
			t.Fatalf("expected 200 OK, got %d: %s", res.StatusCode, string(body))
		}
		var task model.Task
		if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if task.Title != "Updated Empirical Task" || task.Status != model.StatusInProgress || task.Priority != model.PriorityMedium {
			t.Fatalf("unexpected updated fields: %+v", task)
		}
	})

	t.Run("DELETE /api/v1/tasks/{id} -> 204 No Content and subsequent 404", func(t *testing.T) {
		if createdID == "" {
			t.Fatal("no created task ID available")
		}
		req, _ := http.NewRequest(http.MethodDelete, "http://localhost/api/v1/tasks/"+createdID, nil)
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("DELETE error: %v", err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204 No Content, got %d", res.StatusCode)
		}

		// Subsequent GET must return 404
		getRes, err := client.Get("http://localhost/api/v1/tasks/" + createdID)
		if err != nil {
			t.Fatalf("GET after delete error: %v", err)
		}
		defer getRes.Body.Close()
		if getRes.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found after deletion, got %d", getRes.StatusCode)
		}
	})

	// 3. Negative and Boundary Tests
	t.Run("GET /api/v1/tasks/non-existent-id -> 404 Not Found", func(t *testing.T) {
		res, err := client.Get("http://localhost/api/v1/tasks/does-not-exist")
		if err != nil {
			t.Fatalf("GET error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found, got %d", res.StatusCode)
		}
	})

	t.Run("POST empty body -> 400 Bad Request", func(t *testing.T) {
		res, err := client.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(""))
		if err != nil {
			t.Fatalf("POST error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", res.StatusCode)
		}
		b, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(b), "request body must not be empty") {
			t.Fatalf("unexpected message: %s", string(b))
		}
	})

	t.Run("POST empty title -> 400 Bad Request", func(t *testing.T) {
		res, err := client.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(`{"title":"   ","priority":"LOW"}`))
		if err != nil {
			t.Fatalf("POST error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", res.StatusCode)
		}
		b, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(b), "title cannot be empty") {
			t.Fatalf("unexpected error message: %s", string(b))
		}
	})

	t.Run("POST invalid priority -> 400 Bad Request", func(t *testing.T) {
		res, err := client.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(`{"title":"Valid","priority":"SUPER_HIGH"}`))
		if err != nil {
			t.Fatalf("POST error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", res.StatusCode)
		}
		b, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(b), "invalid priority") {
			t.Fatalf("unexpected error message: %s", string(b))
		}
	})

	t.Run("POST malformed JSON -> 400 Bad Request", func(t *testing.T) {
		res, err := client.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(`{"title": "Unclosed`))
		if err != nil {
			t.Fatalf("POST error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", res.StatusCode)
		}
		b, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(b), "malformed JSON") {
			t.Fatalf("unexpected error message: %s", string(b))
		}
	})

	t.Run("POST unknown fields rejected -> 400 Bad Request", func(t *testing.T) {
		res, err := client.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(`{"title":"Valid","priority":"LOW","extra":"bad"}`))
		if err != nil {
			t.Fatalf("POST error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", res.StatusCode)
		}
		b, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(b), "unknown field") || !strings.Contains(string(b), "extra") {
			t.Fatalf("unexpected error message: %s", string(b))
		}
	})

	t.Run("POST multiple JSON objects rejected -> 400 Bad Request", func(t *testing.T) {
		res, err := client.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(`{"title":"A"}{"title":"B"}`))
		if err != nil {
			t.Fatalf("POST error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", res.StatusCode)
		}
		b, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(b), "single JSON object") {
			t.Fatalf("unexpected error message: %s", string(b))
		}
	})

	t.Run("POST payload > 1MB rejected -> 400 Bad Request", func(t *testing.T) {
		largeDesc := strings.Repeat("M", 1024*1024+200)
		body := fmt.Sprintf(`{"title":"Large","description":"%s","priority":"LOW"}`, largeDesc)
		res, err := client.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("POST error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", res.StatusCode)
		}
		b, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(b), "maximum allowed size of 1MB") {
			t.Fatalf("unexpected error message: %s", string(b))
		}
	})

	// 4. Filtering and Search
	t.Run("GET /api/v1/tasks with query filters", func(t *testing.T) {
		_, _ = client.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(`{"title":"Filter Target 1","priority":"HIGH"}`))
		_, _ = client.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(`{"title":"Filter Target 2","priority":"HIGH"}`))
		_, _ = client.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(`{"title":"UniqueSearchKeyword","priority":"LOW"}`))

		// Filter priority=HIGH
		resHigh, err := client.Get("http://localhost/api/v1/tasks?priority=HIGH")
		if err != nil {
			t.Fatalf("GET filter error: %v", err)
		}
		defer resHigh.Body.Close()
		var highList []*model.Task
		if err := json.NewDecoder(resHigh.Body).Decode(&highList); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		for _, task := range highList {
			if task.Priority != model.PriorityHigh {
				t.Fatalf("expected all HIGH, got %s", task.Priority)
			}
		}

		// Search
		resSearch, err := client.Get("http://localhost/api/v1/tasks?search=UniqueSearchKeyword")
		if err != nil {
			t.Fatalf("GET search error: %v", err)
		}
		defer resSearch.Body.Close()
		var searchList []*model.Task
		if err := json.NewDecoder(resSearch.Body).Decode(&searchList); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if len(searchList) != 1 || searchList[0].Title != "UniqueSearchKeyword" {
			t.Fatalf("unexpected search results: %+v", searchList)
		}
	})

	// 5. CORS and Recovery
	t.Run("OPTIONS /api/v1/tasks -> 204 No Content with CORS headers", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodOptions, "http://localhost/api/v1/tasks", nil)
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("OPTIONS error: %v", err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204 No Content, got %d", res.StatusCode)
		}
		if res.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Fatalf("expected CORS origin *, got %s", res.Header.Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("GET /api/v1/panic-trigger -> 500 recovered safely", func(t *testing.T) {
		res, err := client.Get("http://localhost/api/v1/panic-trigger")
		if err != nil {
			t.Fatalf("GET error: %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500 Internal Server Error, got %d", res.StatusCode)
		}
		b, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(b), "internal server error") {
			t.Fatalf("unexpected body: %s", string(b))
		}
	})

	// 6. Concurrency stress test under race detector
	t.Run("Concurrent Stress (50 goroutines x 10 ops under race detector)", func(t *testing.T) {
		numWorkers := 50
		ops := 10
		var wg sync.WaitGroup
		var success int64
		var failure int64

		startBarrier := make(chan struct{})

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				<-startBarrier
				wClient := &http.Client{
					Transport: &handlerTransport{handler: rootHandler},
					Timeout:   5 * time.Second,
				}

				for i := 0; i < ops; i++ {
					if i%2 == 0 {
						p := fmt.Sprintf(`{"title":"Worker %d Task %d","priority":"MEDIUM"}`, workerID, i)
						r, err := wClient.Post("http://localhost/api/v1/tasks", "application/json", strings.NewReader(p))
						if err == nil && r.StatusCode == http.StatusCreated {
							atomic.AddInt64(&success, 1)
						} else {
							atomic.AddInt64(&failure, 1)
						}
						if r != nil {
							io.Copy(io.Discard, r.Body)
							r.Body.Close()
						}
					} else {
						r, err := wClient.Get("http://localhost/api/v1/tasks")
						if err == nil && r.StatusCode == http.StatusOK {
							atomic.AddInt64(&success, 1)
						} else {
							atomic.AddInt64(&failure, 1)
						}
						if r != nil {
							io.Copy(io.Discard, r.Body)
							r.Body.Close()
						}
					}
				}
			}(w)
		}

		close(startBarrier)
		wg.Wait()

		if failure != 0 || success != int64(numWorkers*ops) {
			t.Fatalf("concurrency stress failed: success=%d, failure=%d", success, failure)
		}
	})
}

func TestFileStoragePersistenceIntegrity(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("test_storage_%d.json", time.Now().UnixNano()))
	defer os.Remove(tempFile)

	repo, err := repository.NewFileRepository(tempFile)
	if err != nil {
		t.Fatalf("failed to init file repo: %v", err)
	}

	task := &model.Task{
		ID:        "persist-id-123",
		Title:     "Disk Persistence Test",
		Status:    model.StatusTodo,
		Priority:  model.PriorityHigh,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("failed to create task in file repo: %v", err)
	}

	// Verify file on disk
	info, err := os.Stat(tempFile)
	if err != nil || info.Size() == 0 {
		t.Fatalf("file not written properly: %v", err)
	}

	// Reload in fresh repo instance
	reloadedRepo, err := repository.NewFileRepository(tempFile)
	if err != nil {
		t.Fatalf("failed to reload file repo: %v", err)
	}

	fetched, err := reloadedRepo.GetByID(context.Background(), "persist-id-123")
	if err != nil {
		t.Fatalf("failed to fetch persisted task: %v", err)
	}
	if fetched.Title != "Disk Persistence Test" || fetched.Priority != model.PriorityHigh {
		t.Fatalf("persisted task data corrupted: %+v", fetched)
	}
}
