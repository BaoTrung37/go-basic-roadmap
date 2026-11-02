# Hướng Dẫn Từng Bước Xây Dựng Dự Án TaskManager Bằng Go (Golang)

> **Dành cho người mới bắt đầu**: Hướng dẫn thực hành chi tiết (Step-by-Step Hands-on Tutorial) để tự tay xây dựng từ đầu (from scratch) một dự án backend & CLI hoàn chỉnh theo chuẩn công nghiệp bằng ngôn ngữ Go.
> 
> **Tiêu chí kỹ thuật**:
> - **Ngôn ngữ**: Go 1.22 trở lên.
> - **Phụ thuộc**: **Zero External Dependencies** — 100% sử dụng Thư viện chuẩn (`std`) của Go (`net/http.ServeMux` cải tiến với pattern matching & `r.PathValue`, `log/slog`, `sync.RWMutex`, `sync.WaitGroup`, `context`, `os/signal`, `crypto/rand`, `encoding/json`).
> - **Kiến trúc**: Phân lớp sạch (Clean Layered Architecture), Dependency Injection thủ công, tính đa hình (Polymorphism) qua Interface, và mô hình Concurrency Worker Pool (Goroutines & Channels).

---

## Mục Lục Hướng Dẫn

1. [Tổng Quan Kiến Trúc & Thiết Kế Hệ Thống](#1-tổng-quan-kiến-trúc--thiết-kế-hệ-thống)
2. [Bước 1: Khởi Tạo Go Module & Cấu Trúc Thư Mục Chuẩn](#bước-1-khởi-tạo-go-module--cấu-trúc-thư-mục-chuẩn)
3. [Bước 2: Xây Dựng Domain Models, Enums & Custom Errors (`internal/model`)](#bước-2-xây-dựng-domain-models-enums--custom-errors-internalmodel)
4. [Bước 3: Trừu Tượng Hóa Lưu Trữ với Interface & Cài Đặt Đa Hình (`internal/repository`)](#bước-3-trừu-tượng-hóa-lưu-trữ-với-interface--cài-đặt-đa-hình-internalrepository)
5. [Bước 4: Xây Dựng Concurrency Worker Pool & Tầng Nghiệp Vụ (`internal/worker`, `internal/service`)](#bước-4-xây-dựng-concurrency-worker-pool--tầng-nghiệp-vụ-internalworker-internalservice)
6. [Bước 5: Xây Dựng HTTP Transport Layer & Middlewares Chuẩn Go 1.22+ (`internal/handler`, `internal/middleware`)](#bước-5-xây-dựng-http-transport-layer--middlewares-chuẩn-go-122-internalhandler-internalmiddleware)
7. [Bước 6: Cấu Hình Môi Trường & Xây Dựng Entrypoints (`internal/config`, `cmd/server`, `cmd/cli`)](#bước-6-cấu-hình-môi-trường--xây-dựng-entrypoints-internalconfig-cmdserver-cmdcli)
8. [Bước 7: Viết Unit Tests Toàn Diện & Phát Hiện Race Condition](#bước-7-viết-unit-tests-toàn-diện--phát-hiện-race-condition)
9. [Bước 8: Biên Dịch, Chạy Thử Nghiệm & Thao Tác Toàn Diện](#bước-8-biên-dịch-chạy-thử-nghiệm--thao-tác-toàn-diện)
10. [Tổng Kết & Bài Học Rút Ra](#tổng-kết--bài-học-rút-ra)

---

## 1. Tổng Quan Kiến Trúc & Thiết Kế Hệ Thống

Dự án **TaskManager** được thiết kế nhằm giải quyết bài toán quản lý công việc với 2 giao diện người dùng:
1. **REST API Server** (`cmd/server/main.go`): Cung cấp các endpoint HTTP chuẩn RESTful, tích hợp ghi log có cấu trúc (`log/slog`), bắt lỗi panic (`recovery middleware`), và xử lý sự kiện ngầm bất đồng bộ qua hàng đợi Worker Pool.
2. **Command Line Interface Tool** (`cmd/cli/main.go`): Cho phép người dùng quản lý công việc trực tiếp từ dòng lệnh thông qua các subcommands (`add`, `list`, `get`, `update`, `delete`).

### Sơ Đồ Luồng Dữ Liệu (Data Flow)

```text
       ┌────────────────────────┐         ┌────────────────────────┐
       │   HTTP REST Clients    │         │      CLI Terminal      │
       │     (cURL / Postman)   │         │       (task-cli)       │
       └───────────┬────────────┘         └───────────┬────────────┘
                   │ HTTP Request                     │ CLI Flags/Args
                   ▼                                  ▼
       ┌────────────────────────┐                     │
       │    HTTP Router & Mux   │                     │
       │   (Go 1.22+ ServeMux)  │                     │
       └───────────┬────────────┘                     │
                   │ JSON / Context                   │
                   ▼                                  │
       ┌────────────────────────┐                     │
       │     TaskHandler        │                     │
       │ (internal/handler)     │                     │
       └───────────┬────────────┘                     │
                   │ Call TaskService Methods         │
                   ▼                                  ▼
       ┌───────────────────────────────────────────────────────────┐
       │                       TaskService                         │
       │                  (internal/service)                       │
       │    - Input Validation & Business Rules Enforcement        │
       │    - Event Dispatching to Async Worker Pool               │
       └───────────┬──────────────────────────────────┬────────────┘
                   │                                  │
                   ▼                                  ▼
       ┌────────────────────────┐         ┌────────────────────────┐
       │     TaskRepository     │         │       WorkerPool       │
       │ (internal/repository)  │         │   (internal/worker)    │
       │ ┌────────────────────┐ │         │ - Buffered Channel     │
       │ │ MemoryRepository   │ │         │ - Worker Goroutines    │
       │ ├────────────────────┤ │         │ - Structured Slog      │
       │ │ FileRepository     │ │         │ - Graceful Stop        │
       │ └────────────────────┘ │         └────────────────────────┘
       └───────────┬────────────┘
                   ▼
       ┌────────────────────────┐
       │      Domain Model      │
       │   (internal/model)     │
       │ Task, Status, Priority │
       └────────────────────────┘
```

### Các Nguyên Tắc Thiết Kế Cốt Lõi
- **Đóng gói tầng nội bộ (`internal/`)**: Go cấm các module bên ngoài import các package nằm bên trong thư mục `internal/`. Nhờ đó, mã nguồn cốt lõi được bảo vệ và không bị phụ thuộc ngoài ý muốn.
- **Trừu tượng hóa bằng Interface**: Tầng nghiệp vụ (`service`) chỉ phụ thuộc vào interface `repository.TaskRepository`, không bao giờ gắn chặt (tight coupling) với việc lưu trên RAM hay tệp tin JSON.
- **An toàn Concurrency (Thread-Safety)**: Bộ nhớ trong RAM được bảo vệ bằng `sync.RWMutex`, còn việc ghi đĩa được thực hiện nguyên tử (Atomic write via temporary file rename) nhằm tránh xung đột race condition và ngăn chặn hỏng dữ liệu.

---

## Bước 1: Khởi Tạo Go Module & Cấu Trúc Thư Mục Chuẩn

### 1.1. Khởi Tạo Module

Mở terminal tại thư mục làm việc và khởi tạo Go Module với tên `go-basic-roadmap`:

```bash
mkdir -p go-basic-roadmap
cd go-basic-roadmap
go mod init go-basic-roadmap
```

Lệnh trên sẽ tạo tệp `go.mod` xác định phiên bản Go tối thiểu:

```go
module go-basic-roadmap

go 1.22
```

> **Giải thích**: `go.mod` là tệp định danh module gốc trong hệ sinh thái Go hiện đại. Mọi đường dẫn import nội bộ trong dự án sẽ bắt đầu bằng prefix này (ví dụ: `import "go-basic-roadmap/internal/model"`).

### 1.2. Tạo Khung Thư Mục Chuẩn Công Nghiệp

Thực thi lệnh sau để tạo cây thư mục:

```bash
mkdir -p cmd/server cmd/cli
mkdir -p internal/config internal/model internal/repository internal/service internal/worker internal/handler internal/middleware
mkdir -p docs
```

Ý nghĩa từng thư mục:
- `cmd/`: Chứa các hàm `main` độc lập của từng chương trình thực thi (binary entrypoints). Mỗi thư mục con bên trong `cmd/` tương ứng với một binary output.
- `internal/`: Chứa toàn bộ mã nguồn private của ứng dụng. Trình biên dịch Go bảo vệ nghiêm ngặt thư mục này, không cho phép dự án bên ngoài import.
  - `config/`: Đọc và phân giải cấu hình từ biến môi trường.
  - `model/`: Khai báo các đối tượng nghiệp vụ (Entities), hằng số, quy tắc validation và các lỗi định danh.
  - `repository/`: Tầng lưu trữ dữ liệu với Interface và các cài đặt cụ thể.
  - `service/`: Tầng nghiệp vụ chính (Business Logic) điều phối thao tác.
  - `worker/`: Hệ thống xử lý nền bất đồng bộ (Goroutines & Buffered Channels).
  - `handler/`: Tầng tiếp nhận HTTP Request, điều hướng URL và tuần tự hóa dữ liệu JSON.
  - `middleware/`: Các bộ lọc trung gian xử lý Logging, Panic Recovery, CORS.
- `docs/`: Chứa lộ trình học tập (`roadmap.md`) và tài liệu hướng dẫn (`step-by-step-guide.md`).

---

## Bước 2: Xây Dựng Domain Models, Enums & Custom Errors (`internal/model`)

Domain model là trái tim của hệ thống. Trước khi viết database hay HTTP API, ta cần định nghĩa chính xác dữ liệu mà hệ thống quản lý.

### 2.1. Thiết Kế Enums và Struct Task (`internal/model/task.go`)

Tạo tệp `internal/model/task.go`:

```go
package model

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

// Status đại diện cho trạng thái tiến độ của một công việc.
type Status string

const (
	StatusTodo       Status = "TODO"
	StatusInProgress Status = "IN_PROGRESS"
	StatusDone       Status = "DONE"
)

// Priority đại diện cho mức độ khẩn cấp của công việc.
type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityMedium Priority = "MEDIUM"
	PriorityHigh   Priority = "HIGH"
)

// Task là thực thể nghiệp vụ cốt lõi.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Clone tạo bản sao sâu (deep copy) của Task để ngăn ngừa việc chia sẻ con trỏ gây Data Race.
func (t *Task) Clone() *Task {
	if t == nil {
		return nil
	}
	copied := *t
	return &copied
}

// TaskFilter định nghĩa các tiêu chí lọc danh sách công việc.
type TaskFilter struct {
	Status   Status
	Priority Priority
	Search   string
}

// Các loại sự kiện được phát tới Worker Pool.
const (
	EventTypeCreated       = "CREATED"
	EventTypeUpdated       = "UPDATED"
	EventTypeDeleted       = "DELETED"
	EventTypeStatusChanged = "STATUS_CHANGED"
)

// TaskEvent ghi nhận sự kiện vòng đời của một Task.
type TaskEvent struct {
	Type      string    `json:"type"` // CREATED, UPDATED, DELETED, STATUS_CHANGED
	Task      Task      `json:"task"`
	Timestamp time.Time `json:"timestamp"`
}

// IsValidStatus kiểm tra tính hợp lệ của Status.
func IsValidStatus(s Status) bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// IsValidPriority kiểm tra tính hợp lệ của Priority.
func IsValidPriority(p Priority) bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

// ValidateCreateTask kiểm tra dữ liệu đầu vào khi tạo mới một Task.
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

// ValidateUpdateTask kiểm tra dữ liệu đầu vào khi cập nhật Task.
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

// GenerateID tạo chuỗi định danh duy nhất UUID v4 bằng thư viện chuẩn crypto/rand.
func GenerateID() (string, error) {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		return "", err
	}

	// Đặt version 4 (0100) vào 4 bit cao của byte thứ 6
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// Đặt variant 10 vào 2 bit cao của byte thứ 8
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}
```

### 2.2. Phân Cấp Lỗi Nghiệp Vụ Định Danh (`internal/model/errors.go`)

Trong Go, cách tiếp cận chuẩn mực là sử dụng **Sentinel Errors** kết hợp **Error Wrapping (`%w`)**. Tạo tệp `internal/model/errors.go`:

```go
package model

import "errors"

// Các lỗi định danh mức Domain của TaskManager.
var (
	// ErrTaskNotFound trả về khi không tìm thấy công việc với ID tương ứng.
	ErrTaskNotFound = errors.New("task not found")

	// ErrInvalidInput trả về khi dữ liệu đầu vào không hợp lệ.
	ErrInvalidInput = errors.New("invalid input")

	// ErrTaskConflict trả về khi ID công việc đã tồn tại trong bộ lưu trữ.
	ErrTaskConflict = errors.New("task conflict: task with given ID already exists")

	// ErrStorageFailure trả về khi thao tác I/O lưu trữ tệp tin gặp sự cố.
	ErrStorageFailure = errors.New("storage failure")

	// Các lỗi chi tiết của quá trình kiểm tra hợp lệ
	ErrTitleRequired   = errors.New("title cannot be empty")
	ErrTitleTooLong    = errors.New("title must not exceed 100 characters")
	ErrInvalidStatus   = errors.New("invalid status: must be TODO, IN_PROGRESS, or DONE")
	ErrInvalidPriority = errors.New("invalid priority: must be LOW, MEDIUM, or HIGH")
)
```

> **Điểm cốt lõi cần nhớ**:
> 1. Sử dụng `errors.Is(err, model.ErrTaskNotFound)` để kiểm tra lỗi ở tầng trên (HTTP Handler) thay vì so sánh chuỗi bằng `err.Error() == "..."`.
> 2. `GenerateID()` sinh chuỗi UUID v4 hoàn toàn bằng `crypto/rand` của Go Standard Library, không cần kéo dependency `github.com/google/uuid`.

---

## Bước 3: Trừu Tượng Hóa Lưu Trữ với Interface & Cài Đặt Đa Hình (`internal/repository`)

### 3.1. Định Nghĩa Interface `TaskRepository` (`internal/repository/repository.go`)

Tạo tệp `internal/repository/repository.go`:

```go
package repository

import (
	"context"
	"go-basic-roadmap/internal/model"
)

// TaskRepository trừu tượng hóa toàn bộ hành vi lưu trữ dữ liệu.
// Mọi phương thức đều nhận context.Context để hỗ trợ hủy bỏ thao tác khi timeout/cancel.
type TaskRepository interface {
	// Create lưu một công việc mới. Trả về ErrTaskConflict nếu ID đã tồn tại.
	Create(ctx context.Context, task *model.Task) error

	// GetByID lấy thông tin công việc theo ID. Trả về ErrTaskNotFound nếu không tồn tại.
	GetByID(ctx context.Context, id string) (*model.Task, error)

	// List trả về danh sách công việc thỏa mãn bộ lọc.
	List(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error)

	// Update cập nhật công việc hiện có. Trả về ErrTaskNotFound nếu không tìm thấy.
	Update(ctx context.Context, task *model.Task) error

	// Delete xóa công việc theo ID. Trả về ErrTaskNotFound nếu không tìm thấy.
	Delete(ctx context.Context, id string) error
}
```

### 3.2. Cài Đặt Bộ Nhớ Trong Thread-Safe (`internal/repository/memory.go`)

Để đảm bảo hiệu năng cao và thuận tiện khi chạy thử nghiệm hoặc viết unit test, ta tạo `MemoryRepository` sử dụng `sync.RWMutex`:

```go
package repository

import (
	"context"
	"sort"
	"strings"
	"sync"

	"go-basic-roadmap/internal/model"
)

// MemoryRepository triển khai TaskRepository trên RAM với cơ chế khóa đọc/ghi.
type MemoryRepository struct {
	mu    sync.RWMutex
	tasks map[string]*model.Task
}

// NewMemoryRepository khởi tạo một thể hiện MemoryRepository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tasks: make(map[string]*model.Task),
	}
}

func (r *MemoryRepository) Create(ctx context.Context, task *model.Task) error {
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

	// Defensive copy: Luôn Clone dữ liệu trước khi lưu vào map
	r.tasks[task.ID] = task.Clone()
	return nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (*model.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, model.ErrTaskNotFound
	}

	// Trả về bản sao, bảo vệ dữ liệu trong map không bị sửa đổi ngoài ý muốn
	return task.Clone(), nil
}

func (r *MemoryRepository) List(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
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

	// Sắp xếp giảm dần theo thời gian tạo (mới nhất lên đầu)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	return results, nil
}

func (r *MemoryRepository) Update(ctx context.Context, task *model.Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if task == nil {
		return model.ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return model.ErrTaskNotFound
	}

	r.tasks[task.ID] = task.Clone()
	return nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return model.ErrTaskNotFound
	}

	delete(r.tasks, id)
	return nil
}
```

### 3.3. Cài Đặt Lưu Trữ Tệp Tin Nguyên Tử (`internal/repository/file.go`)

Để dữ liệu không bị mất khi ứng dụng tắt, ta xây dựng `FileRepository`. Điểm đặc biệt là cơ chế **Atomic File Write**:
- Ghi dữ liệu ra tệp tạm `.tmp`.
- Dùng `os.Rename(tmpFile, finalFile)` để thay thế nguyên tử. Hệ điều hành đảm bảo tiến trình đổi tên không bao giờ bị cắt ngang giữa chừng, bảo vệ tệp tin JSON không bao giờ bị rỗng hay hỏng khi gặp sự cố mất điện.

Tạo tệp `internal/repository/file.go`:

```go
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

type FileRepository struct {
	mu       sync.RWMutex
	filePath string
	tasks    map[string]*model.Task
}

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

func (r *FileRepository) load() error {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
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

func (r *FileRepository) save() error {
	taskList := make([]*model.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		taskList = append(taskList, task)
	}

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

	// Thay thế nguyên tử bằng os.Rename
	if err := os.Rename(tmpFile, r.filePath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("%w: failed to atomically replace data file: %v", model.ErrStorageFailure, err)
	}

	return nil
}

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
		delete(r.tasks, task.ID) // Roll back khi ghi đĩa thất bại
		return err
	}

	return nil
}

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
		r.tasks[task.ID] = oldTask // Roll back
		return err
	}

	return nil
}

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
		r.tasks[id] = oldTask // Roll back
		return err
	}

	return nil
}
```

---

## Bước 4: Xây Dựng Concurrency Worker Pool & Tầng Nghiệp Vụ (`internal/worker`, `internal/service`)

### 4.1. Concurrency Worker Pool (`internal/worker/worker_pool.go`)

Để xử lý các tác vụ nền như ghi log kiểm toán (Audit Log) hoặc gửi thông báo bất đồng bộ mà không chặn (block) request của người dùng, ta sử dụng mẫu thiết kế **Worker Pool**.

Tạo tệp `internal/worker/worker_pool.go`:

```go
package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"go-basic-roadmap/internal/model"
)

type WorkerPool interface {
	Start(ctx context.Context)
	Dispatch(event model.TaskEvent) bool
	Stop()
}

type Pool struct {
	mu         sync.RWMutex
	numWorkers int
	capacity   int
	queue      chan model.TaskEvent
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *slog.Logger
	started    bool
	stopped    bool
	stopOnce   sync.Once
}

func NewWorkerPool(numWorkers int, bufferCapacity int, logger *slog.Logger) *Pool {
	if numWorkers <= 0 {
		numWorkers = 1
	}
	if bufferCapacity <= 0 {
		bufferCapacity = 10
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Pool{
		numWorkers: numWorkers,
		capacity:   bufferCapacity,
		queue:      make(chan model.TaskEvent, bufferCapacity),
		logger:     logger,
	}
}

func (p *Pool) Start(ctx context.Context) {
	p.mu.Lock()
	if p.started || p.stopped {
		p.mu.Unlock()
		return
	}
	p.started = true
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.mu.Unlock()

	p.logger.Info("Starting worker pool", "workers", p.numWorkers, "buffer_capacity", p.capacity)

	for i := 1; i <= p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// Dispatch đưa sự kiện vào hàng đợi. Nếu hàng đợi đầy, bỏ qua và trả về false (Non-blocking).
func (p *Pool) Dispatch(event model.TaskEvent) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started || p.stopped {
		return false
	}

	select {
	case p.queue <- event:
		return true
	default:
		p.logger.Warn("Worker event queue is full, dropping event",
			"event_type", event.Type,
			"task_id", event.Task.ID,
		)
		return false
	}
}

// Stop dừng pool an toàn: đóng channel, chờ các worker xử lý hết sự kiện rồi thoát.
func (p *Pool) Stop() {
	p.stopOnce.Do(func() {
		p.logger.Info("Stopping worker pool...")

		p.mu.Lock()
		p.stopped = true
		close(p.queue)
		p.mu.Unlock()

		p.wg.Wait()

		if p.cancel != nil {
			p.cancel()
		}

		p.logger.Info("Worker pool stopped cleanly")
	})
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			// Khi context bị hủy, hút cạn (drain) các item còn lại trong hàng đợi
			for {
				select {
				case event, ok := <-p.queue:
					if !ok {
						return
					}
					p.processEvent(id, event)
				default:
					return
				}
			}
		case event, ok := <-p.queue:
			if !ok {
				return
			}
			p.processEvent(id, event)
		}
	}
}

func (p *Pool) processEvent(workerID int, event model.TaskEvent) {
	p.logger.Info("Worker processed task event",
		"worker_id", workerID,
		"event_type", event.Type,
		"task_id", event.Task.ID,
		"status", event.Task.Status,
		"priority", event.Task.Priority,
		"timestamp", event.Timestamp.Format(time.RFC3339),
	)
}

func (p *Pool) QueueLength() int {
	return len(p.queue)
}

func (p *Pool) Capacity() int {
	return p.capacity
}
```

### 4.2. Tầng Nghiệp Vụ TaskService (`internal/service/task_service.go`)

Tạo tệp `internal/service/task_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/repository"
	"go-basic-roadmap/internal/worker"
)

type TaskService interface {
	CreateTask(ctx context.Context, title, description string, priority model.Priority) (*model.Task, error)
	GetTask(ctx context.Context, id string) (*model.Task, error)
	ListTasks(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error)
	UpdateTask(ctx context.Context, id, title, description string, status model.Status, priority model.Priority) (*model.Task, error)
	DeleteTask(ctx context.Context, id string) error
}

type taskService struct {
	repo   repository.TaskRepository
	pool   worker.WorkerPool
	logger *slog.Logger
}

func NewTaskService(repo repository.TaskRepository, pool worker.WorkerPool, logger *slog.Logger) TaskService {
	if logger == nil {
		logger = slog.Default()
	}
	return &taskService{
		repo:   repo,
		pool:   pool,
		logger: logger,
	}
}

func (s *taskService) CreateTask(ctx context.Context, title, description string, priority model.Priority) (*model.Task, error) {
	if priority == "" {
		priority = model.PriorityMedium
	}

	if err := model.ValidateCreateTask(title, priority); err != nil {
		return nil, err
	}

	id, err := model.GenerateID()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate task ID: %v", model.ErrStorageFailure, err)
	}

	now := time.Now().UTC()
	task := &model.Task{
		ID:          id,
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		Status:      model.StatusTodo,
		Priority:    priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	if s.pool != nil {
		s.pool.Dispatch(model.TaskEvent{
			Type:      model.EventTypeCreated,
			Task:      *task,
			Timestamp: now,
		})
	}

	return task, nil
}

func (s *taskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, fmt.Errorf("%w: task ID cannot be empty", model.ErrInvalidInput)
	}
	return s.repo.GetByID(ctx, cleanID)
}

func (s *taskService) ListTasks(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
	return s.repo.List(ctx, filter)
}

func (s *taskService) UpdateTask(ctx context.Context, id, title, description string, status model.Status, priority model.Priority) (*model.Task, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, fmt.Errorf("%w: task ID cannot be empty", model.ErrInvalidInput)
	}

	if err := model.ValidateUpdateTask(title, status, priority); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, cleanID)
	if err != nil {
		return nil, err
	}

	statusChanged := (existing.Status != status)
	now := time.Now().UTC()

	updated := &model.Task{
		ID:          cleanID,
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		Status:      status,
		Priority:    priority,
		CreatedAt:   existing.CreatedAt,
		UpdatedAt:   now,
	}

	if err := s.repo.Update(ctx, updated); err != nil {
		return nil, err
	}

	if s.pool != nil {
		eventType := model.EventTypeUpdated
		if statusChanged {
			eventType = model.EventTypeStatusChanged
		}
		s.pool.Dispatch(model.TaskEvent{
			Type:      eventType,
			Task:      *updated,
			Timestamp: now,
		})
	}

	return updated, nil
}

func (s *taskService) DeleteTask(ctx context.Context, id string) error {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return fmt.Errorf("%w: task ID cannot be empty", model.ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, cleanID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, cleanID); err != nil {
		return err
	}

	if s.pool != nil {
		s.pool.Dispatch(model.TaskEvent{
			Type:      model.EventTypeDeleted,
			Task:      *existing,
			Timestamp: time.Now().UTC(),
		})
	}

	return nil
}
```

---

## Bước 5: Xây Dựng HTTP Transport Layer & Middlewares Chuẩn Go 1.22+ (`internal/handler`, `internal/middleware`)

### 5.1. Tiện Ích JSON & Định Dạng Phản Hồi (`internal/handler/response.go`)

Tạo tệp `internal/handler/response.go` để đóng gói chuẩn mã hóa/giải mã JSON an toàn:

```go
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, `{"error":"failed to encode JSON response"}`, http.StatusInternalServerError)
		}
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{
		Error: message,
		Code:  status,
	})
}

// ReadJSON giải mã JSON an toàn: giới hạn tối đa 1MB và từ chối trường không xác định.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := int64(1048576) // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("malformed JSON at position %d", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("malformed JSON: unexpected end of body")
		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("invalid value for field %q at position %d", unmarshalTypeError.Field, unmarshalTypeError.Offset)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("unknown field %s in JSON body", fieldName)
		case errors.Is(err, io.EOF):
			return errors.New("request body must not be empty")
		case err.Error() == "http: request body too large":
			return errors.New("request body exceeds maximum allowed size of 1MB")
		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain only a single JSON object")
	}

	return nil
}
```

### 5.2. HTTP Handler Đăng Ký Route (`internal/handler/task_handler.go`)

Từ Go 1.22+, `http.ServeMux` hỗ trợ:
- Khai báo phương thức HTTP ngay trước pattern: `"POST /api/v1/tasks"`
- Biến đường dẫn: `"GET /api/v1/tasks/{id}"`
- Đọc biến bằng `r.PathValue("id")`

Tạo tệp `internal/handler/task_handler.go`:

```go
package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/service"
)

type CreateTaskRequest struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Priority    model.Priority `json:"priority"`
}

type UpdateTaskRequest struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      model.Status   `json:"status"`
	Priority    model.Priority `json:"priority"`
}

type TaskHandler struct {
	service service.TaskService
	logger  *slog.Logger
}

func NewTaskHandler(svc service.TaskService, logger *slog.Logger) *TaskHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &TaskHandler{
		service: svc,
		logger:  logger,
	}
}

func (h *TaskHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tasks", h.CreateTask)
	mux.HandleFunc("GET /api/v1/tasks", h.ListTasks)
	mux.HandleFunc("GET /api/v1/tasks/{id}", h.GetTask)
	mux.HandleFunc("PUT /api/v1/tasks/{id}", h.UpdateTask)
	mux.HandleFunc("DELETE /api/v1/tasks/{id}", h.DeleteTask)
	mux.HandleFunc("GET /health", h.HealthCheck)
}

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

func (h *TaskHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
```

### 5.3. Xây Dựng Middlewares (`internal/middleware/middleware.go`)

Tạo tệp `internal/middleware/middleware.go` gồm 3 bộ lọc quan trọng:
1. `Logging`: Sử dụng `log/slog` ghi lại Method, Path, Status Code, và Duration (ms).
2. `Recovery`: Bắt mọi `panic()` trong Handler và trả về 500 thay vì làm rớt server.
3. `CORS`: Cung cấp tiêu đề HTTP cho phép trình duyệt truy cập từ tên miền khác.

```go
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriterInterceptor) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriterInterceptor) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			interceptor := &responseWriterInterceptor{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(interceptor, r)

			duration := time.Since(start)
			logger.Info("HTTP request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", interceptor.statusCode,
				"duration_ms", duration.Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"bytes", interceptor.bytesWritten,
			)
		})
	}
}

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					logger.Error("Panic recovered in HTTP handler",
						"panic", rec,
						"path", r.URL.Path,
						"stack", string(stack),
					)

					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"internal server error","code":500}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

---

## Bước 6: Cấu Hình Môi Trường & Xây Dựng Entrypoints (`internal/config`, `cmd/server`, `cmd/cli`)

### 6.1. Quản Lý Cấu Hình Môi Trường (`internal/config/config.go`)

Tạo tệp `internal/config/config.go`:

```go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort      string
	StorageType     string // "memory" hoặc "file"
	DataFilePath    string
	WorkerPoolSize  int
	WorkerQueueSize int
	LogLevel        string
}

func Load() *Config {
	port := getEnv("SERVER_PORT", "")
	if port == "" {
		port = getEnv("PORT", ":8080")
	}
	if len(port) > 0 && port[0] != ':' {
		port = ":" + port
	}

	return &Config{
		ServerPort:      port,
		StorageType:     getEnv("STORAGE_TYPE", "memory"),
		DataFilePath:    getEnv("DATA_FILE_PATH", "data/tasks.json"),
		WorkerPoolSize:  getEnvInt("WORKER_POOL_SIZE", 3),
		WorkerQueueSize: getEnvInt("WORKER_QUEUE_SIZE", 100),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if intVal, err := strconv.Atoi(val); err == nil && intVal > 0 {
			return intVal
		}
	}
	return defaultVal
}
```

### 6.2. Server Entrypoint Với Graceful Shutdown (`cmd/server/main.go`)

Tạo tệp `cmd/server/main.go`. Quy trình khởi tạo:
1. Khởi tạo `slog.NewJSONHandler` phục vụ ghi log theo định dạng JSON chuẩn.
2. Tải cấu hình từ `config.Load()`.
3. Khởi tạo `TaskRepository` (`Memory` hoặc `File` dựa theo `STORAGE_TYPE`).
4. Khởi động `WorkerPool` trên background context.
5. Khởi tạo `TaskService` và `TaskHandler`.
6. Bọc các middleware theo thứ tự: `Recovery` -> `Logging` -> `CORS`.
7. Lắng nghe tín hiệu hệ thống (`os.Interrupt`, `syscall.SIGTERM`) và thực hiện Graceful Shutdown.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-basic-roadmap/internal/config"
	"go-basic-roadmap/internal/handler"
	"go-basic-roadmap/internal/middleware"
	"go-basic-roadmap/internal/repository"
	"go-basic-roadmap/internal/service"
	"go-basic-roadmap/internal/worker"
)

func main() {
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	cfg := config.Load()
	logger.Info("Starting TaskManager server",
		"port", cfg.ServerPort,
		"storage_type", cfg.StorageType,
		"worker_pool_size", cfg.WorkerPoolSize,
	)

	// 1. Khởi tạo Repository
	var repo repository.TaskRepository
	switch cfg.StorageType {
	case "file":
		fileRepo, err := repository.NewFileRepository(cfg.DataFilePath)
		if err != nil {
			logger.Error("Failed to initialize file repository", "path", cfg.DataFilePath, "error", err)
			os.Exit(1)
		}
		repo = fileRepo
		logger.Info("Initialized FileTaskRepository", "path", cfg.DataFilePath)
	default:
		repo = repository.NewMemoryRepository()
		logger.Info("Initialized MemoryTaskRepository")
	}

	// 2. Khởi tạo Worker Pool
	workerPool := worker.NewWorkerPool(cfg.WorkerPoolSize, cfg.WorkerQueueSize, logger)
	poolCtx, cancelPool := context.WithCancel(context.Background())
	defer cancelPool()
	workerPool.Start(poolCtx)

	// 3. Khởi tạo Service
	taskService := service.NewTaskService(repo, workerPool, logger)

	// 4. Khởi tạo Router HTTP
	mux := http.NewServeMux()
	taskHandler := handler.NewTaskHandler(taskService, logger)
	taskHandler.RegisterRoutes(mux)

	// 5. Áp dụng Middlewares: CORS -> Recovery -> Logging
	var rootHandler http.Handler = mux
	rootHandler = middleware.Recovery(logger)(rootHandler)
	rootHandler = middleware.Logging(logger)(rootHandler)
	rootHandler = middleware.CORS(rootHandler)

	server := &http.Server{
		Addr:         cfg.ServerPort,
		Handler:      rootHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 6. Chạy Server trên Goroutine nền
	serverErrChan := make(chan error, 1)
	go func() {
		logger.Info(fmt.Sprintf("HTTP server listening on %s", cfg.ServerPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	// 7. Chờ tín hiệu dừng từ Hệ điều hành
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		logger.Error("HTTP server failed to start or run", "error", err)
		workerPool.Stop()
		os.Exit(1)
	case sig := <-shutdownChan:
		logger.Info("Received termination signal, starting graceful shutdown", "signal", sig.String())
	}

	// 8. Graceful Shutdown với Timeout 10 giây
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server graceful shutdown failed, forcing close", "error", err)
		_ = server.Close()
	} else {
		logger.Info("HTTP server stopped gracefully")
	}

	workerPool.Stop()
	logger.Info("TaskManager server exited cleanly")
}
```

### 6.3. CLI Entrypoint (`cmd/cli/main.go`)

Tạo tệp `cmd/cli/main.go` triển khai công cụ CLI:

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"go-basic-roadmap/internal/config"
	"go-basic-roadmap/internal/model"
	"go-basic-roadmap/internal/repository"
	"go-basic-roadmap/internal/service"
)

func main() {
	cfg := config.Load()
	dataPath := flag.String("data", cfg.DataFilePath, "Path to JSON tasks data file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "TaskManager CLI - Pure Go Task Management Tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  task-cli [flags] <command> [arguments]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  add     Create a new task\n")
		fmt.Fprintf(os.Stderr, "  list    List tasks with optional filters\n")
		fmt.Fprintf(os.Stderr, "  get     Retrieve task details by ID\n")
		fmt.Fprintf(os.Stderr, "  update  Update an existing task\n")
		fmt.Fprintf(os.Stderr, "  delete  Remove a task by ID\n\n")
		fmt.Fprintf(os.Stderr, "Global Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	repo, err := repository.NewFileRepository(*dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing repository: %v\n", err)
		os.Exit(1)
	}

	svc := service.NewTaskService(repo, nil, nil)
	ctx := context.Background()

	command := args[0]
	cmdArgs := args[1:]

	switch command {
	case "add":
		handleAdd(ctx, svc, cmdArgs)
	case "list":
		handleList(ctx, svc, cmdArgs)
	case "get":
		handleGet(ctx, svc, cmdArgs)
	case "update":
		handleUpdate(ctx, svc, cmdArgs)
	case "delete":
		handleDelete(ctx, svc, cmdArgs)
	case "help":
		flag.Usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func handleAdd(ctx context.Context, svc service.TaskService, args []string) {
	cmd := flag.NewFlagSet("add", flag.ExitOnError)
	title := cmd.String("title", "", "Task title (required)")
	desc := cmd.String("desc", "", "Task description")
	priority := cmd.String("priority", "MEDIUM", "Priority: LOW, MEDIUM, HIGH")
	_ = cmd.Parse(args)

	if strings.TrimSpace(*title) == "" {
		fmt.Fprintln(os.Stderr, "Error: --title is required")
		cmd.Usage()
		os.Exit(1)
	}

	task, err := svc.CreateTask(ctx, *title, *desc, model.Priority(strings.ToUpper(*priority)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create task: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task created successfully!\n")
	printTaskDetails(task)
}

func handleList(ctx context.Context, svc service.TaskService, args []string) {
	cmd := flag.NewFlagSet("list", flag.ExitOnError)
	status := cmd.String("status", "", "Filter by status: TODO, IN_PROGRESS, DONE")
	priority := cmd.String("priority", "", "Filter by priority: LOW, MEDIUM, HIGH")
	search := cmd.String("search", "", "Search keywords in title or description")
	_ = cmd.Parse(args)

	filter := model.TaskFilter{
		Status:   model.Status(strings.ToUpper(*status)),
		Priority: model.Priority(strings.ToUpper(*priority)),
		Search:   *search,
	}

	tasks, err := svc.ListTasks(ctx, filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list tasks: %v\n", err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found matching criteria.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tPRIORITY\tTITLE\tCREATED")
	fmt.Fprintln(w, "--\t------\t--------\t-----\t-------")
	for _, t := range tasks {
		created := t.CreatedAt.Local().Format("2006-01-02 15:04")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.ID, t.Status, t.Priority, t.Title, created)
	}
	_ = w.Flush()
}

func handleGet(ctx context.Context, svc service.TaskService, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: task-cli get <task-id>")
		os.Exit(1)
	}

	id := args[0]
	task, err := svc.GetTask(ctx, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error retrieving task: %v\n", err)
		os.Exit(1)
	}

	printTaskDetails(task)
}

func handleUpdate(ctx context.Context, svc service.TaskService, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: task-cli update <task-id> [flags]")
		os.Exit(1)
	}

	id := args[0]
	cmd := flag.NewFlagSet("update", flag.ExitOnError)
	title := cmd.String("title", "", "Updated title")
	desc := cmd.String("desc", "", "Updated description")
	status := cmd.String("status", "", "Updated status (TODO, IN_PROGRESS, DONE)")
	priority := cmd.String("priority", "", "Updated priority (LOW, MEDIUM, HIGH)")
	_ = cmd.Parse(args[1:])

	existing, err := svc.GetTask(ctx, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error retrieving task to update: %v\n", err)
		os.Exit(1)
	}

	newTitle := existing.Title
	if *title != "" {
		newTitle = *title
	}
	newDesc := existing.Description
	if *desc != "" {
		newDesc = *desc
	}
	newStatus := existing.Status
	if *status != "" {
		newStatus = model.Status(strings.ToUpper(*status))
	}
	newPriority := existing.Priority
	if *priority != "" {
		newPriority = model.Priority(strings.ToUpper(*priority))
	}

	updated, err := svc.UpdateTask(ctx, id, newTitle, newDesc, newStatus, newPriority)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to update task: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task updated successfully!\n")
	printTaskDetails(updated)
}

func handleDelete(ctx context.Context, svc service.TaskService, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: task-cli delete <task-id>")
		os.Exit(1)
	}

	id := args[0]
	err := svc.DeleteTask(ctx, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to delete task: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task %q deleted successfully.\n", id)
}

func printTaskDetails(task *model.Task) {
	fmt.Printf("ID:          %s\n", task.ID)
	fmt.Printf("Title:       %s\n", task.Title)
	if task.Description != "" {
		fmt.Printf("Description: %s\n", task.Description)
	}
	fmt.Printf("Status:      %s\n", task.Status)
	fmt.Printf("Priority:    %s\n", task.Priority)
	fmt.Printf("Created:     %s\n", task.CreatedAt.Local().Format(time.RFC1123))
	fmt.Printf("Updated:     %s\n", task.UpdatedAt.Local().Format(time.RFC1123))
}
```

---

## Bước 7: Viết Unit Tests Toàn Diện & Phát Hiện Race Condition

Go có triết lý kiểm thử xuất sắc tích hợp sẵn trong công cụ `go test`. Không cần bất kỳ thư viện Assertions hay Mocking phức tạp nào bên ngoài, ta sử dụng:
1. **Table-Driven Tests**: Định nghĩa mảng các test cases chứa input và output mong đợi.
2. **`net/http/httptest`**: Giả lập request và response HTTP trong bộ nhớ siêu nhanh.
3. **`-race` Detector**: Cờ biên dịch kiểm tra xung đột truy cập bộ nhớ đa luồng.

### 7.1. Cấu Trúc Kiểm Thử Bảng (Table-Driven Test)

Ví dụ kiểm thử validation trong `internal/model/task_test.go`:

```go
func TestValidateCreateTask(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		priority    model.Priority
		expectErr   bool
		expectedErr error
	}{
		{
			name:      "hợp lệ với tiêu đề và độ ưu tiên mặc định",
			title:     "Học Go Concurrency",
			priority:  model.PriorityHigh,
			expectErr: false,
		},
		{
			name:        "lỗi khi tiêu đề bị để trống",
			title:       "   ",
			priority:    model.PriorityMedium,
			expectErr:   true,
			expectedErr: model.ErrTitleRequired,
		},
		{
			name:        "lỗi khi tiêu đề vượt quá 100 ký tự",
			title:       strings.Repeat("a", 101),
			priority:    model.PriorityLow,
			expectErr:   true,
			expectedErr: model.ErrTitleTooLong,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := model.ValidateCreateTask(tc.title, tc.priority)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("kỳ vọng phát sinh lỗi nhưng kết quả trả về nil")
				}
				if tc.expectedErr != nil && !errors.Is(err, tc.expectedErr) {
					t.Errorf("kỳ vọng lỗi %v, nhận được: %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("kỳ vọng thành công nhưng gặp lỗi: %v", err)
			}
		})
	}
}
```

### 7.2. Kiểm Thử HTTP Handler với `httptest`

Trong `internal/handler/task_handler_test.go`:

```go
func TestCreateTaskHandler(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewTaskService(repo, nil, nil)
	h := handler.NewTaskHandler(svc, nil)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	payload := `{"title":"Làm bài tập Go","description":"Chi tiết bước 1 đến 8","priority":"HIGH"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("kỳ vọng mã 201 Created, nhận được: %d, body: %s", w.Code, w.Body.String())
	}
}
```

### 7.3. Lệnh Chạy Kiểm Thử & Kiểm Tra Race Condition

```bash
# 1. Chạy toàn bộ unit tests với log chi tiết
go test -v ./...

# 2. Chạy kiểm thử kết hợp công cụ phát hiện xung đột bộ nhớ (-race)
go test -v -race ./...

# 3. Phân tích độ bao phủ mã nguồn (Coverage)
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

---

## Bước 8: Biên Dịch, Chạy Thử Nghiệm & Thao Tác Toàn Diện

### 8.1. Biên Dịch Dự Án Bằng Makefile

Dự án cung cấp `Makefile` với đầy đủ các mục tiêu tự động hóa:

```bash
# Xem danh sách các lệnh hỗ trợ
make help

# Biên dịch cả Server và CLI
make build
```

Sau khi chạy `make build`, thư mục `bin/` sẽ chứa 2 tệp thực thi:
- `bin/server`
- `bin/task-cli`

### 8.2. Khởi Chạy REST API Server

```bash
# Chạy trực tiếp từ mã nguồn:
make run-server

# Hoặc chạy với biến môi trường tùy chỉnh:
STORAGE_TYPE=file DATA_FILE_PATH=data/my_tasks.json PORT=8080 ./bin/server
```

### 8.3. Thao Tác Toàn Bộ API Bằng cURL

Mở một tab terminal khác và chạy các lệnh kiểm thử sau:

#### 1. Kiểm tra trạng thái máy chủ (Health Check):
```bash
curl -i http://localhost:8080/health
```
*Phản hồi mẫu:*
```json
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8

{
  "status": "healthy",
  "timestamp": "2026-11-02T02:30:00Z"
}
```

#### 2. Tạo công việc mới (Create Task):
```bash
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Học lập trình Go chuẩn công nghiệp","description":"Thực hành xây dựng TaskManager REST API và CLI","priority":"HIGH"}'
```
*Phản hồi mẫu (HTTP 201 Created):*
```json
HTTP/1.1 201 Created
Content-Type: application/json; charset=utf-8

{
  "id": "e7c65349-f001-4475-8160-520e58d7c229",
  "title": "Học lập trình Go chuẩn công nghiệp",
  "description": "Thực hành xây dựng TaskManager REST API và CLI",
  "status": "TODO",
  "priority": "HIGH",
  "created_at": "2026-11-02T02:30:15Z",
  "updated_at": "2026-11-02T02:30:15Z"
}
```

#### 3. Lấy danh sách công việc (List Tasks):
```bash
curl -i http://localhost:8080/api/v1/tasks
```

#### 4. Lọc công việc theo tiêu chí (Filter by Status, Priority, Search):
```bash
# Lọc theo Status TODO và Priority HIGH:
curl -i "http://localhost:8080/api/v1/tasks?status=TODO&priority=HIGH"

# Tìm kiếm từ khóa trong title/description:
curl -i "http://localhost:8080/api/v1/tasks?search=TaskManager"
```

#### 5. Xem chi tiết công việc theo ID (Get Task by ID):
```bash
# Thay thế <TASK_ID> bằng ID thực tế được tạo ở bước 2:
curl -i http://localhost:8080/api/v1/tasks/<TASK_ID>
```

#### 6. Cập nhật công việc (Update Task):
```bash
curl -i -X PUT http://localhost:8080/api/v1/tasks/<TASK_ID> \
  -H "Content-Type: application/json" \
  -d '{"title":"Học lập trình Go chuẩn công nghiệp","description":"Đã hoàn thành bước 1 đến bước 8","status":"IN_PROGRESS","priority":"HIGH"}'
```

#### 7. Xóa công việc (Delete Task):
```bash
curl -i -X DELETE http://localhost:8080/api/v1/tasks/<TASK_ID>
```
*Phản hồi mẫu (HTTP 204 No Content):*
```text
HTTP/1.1 204 No Content
```

### 8.4. Thao Tác Bằng Công Cụ Dòng Lệnh CLI (`task-cli`)

Chạy công cụ CLI tương tác với tệp dữ liệu JSON:

```bash
# 1. Thêm mới công việc
./bin/task-cli -data data/tasks.json add -title "Đọc tài liệu Concurrency" -desc "Hiểu rõ Channel và Mutex" -priority HIGH

# 2. Xem danh sách công việc dưới dạng bảng
./bin/task-cli -data data/tasks.json list

# 3. Lọc danh sách theo trạng thái và tìm kiếm
./bin/task-cli -data data/tasks.json list -status TODO -search "Concurrency"

# 4. Xem chi tiết công việc
./bin/task-cli -data data/tasks.json get <TASK_ID>

# 5. Cập nhật trạng thái công việc sang DONE
./bin/task-cli -data data/tasks.json update <TASK_ID> -status DONE

# 6. Xóa công việc
./bin/task-cli -data data/tasks.json delete <TASK_ID>
```

---

## 10. Tổng Kết & Bài Học Rút Ra

Qua quá trình tự tay xây dựng dự án **TaskManager**, bạn đã nắm vững các kỹ năng và tư duy cốt lõi của một Go developer thực thụ:

1. **Làm chủ Thư viện chuẩn (`std`)**: Bạn đã thấy sức mạnh to lớn của Go 1.22+ khi có thể xây dựng REST API hoàn chỉnh, routing tham số động, ghi log JSON, và quản lý concurrency mà không cần bất kỳ framework bên ngoài nào.
2. **Tư duy Kiến trúc sạch (Clean Architecture)**: Bằng cách tách biệt Interface `TaskRepository` khỏi cài đặt cụ thể (`MemoryRepository`, `FileRepository`), việc thay đổi nơi lưu trữ (sau này có thể là PostgreSQL hay MongoDB) hoàn toàn không ảnh hưởng đến tầng nghiệp vụ `TaskService`.
3. **Kỹ thuật Concurrency an toàn**: Bạn đã hiểu sâu sắc cách phối hợp `sync.RWMutex` để bảo vệ tài nguyên chia sẻ trong RAM, sử dụng cơ chế ghi đĩa nguyên tử `Atomic File Write` chống corrupt dữ liệu, và vận hành mô hình Worker Pool (`chan`, `goroutine`, `sync.WaitGroup`, `select ... default`) để xử lý tác vụ nền.
4. **Quy trình Graceful Shutdown**: Nắm vững cách bắt tín hiệu hệ điều hành (`os.Signal`) và hủy bỏ context để ứng dụng tắt mượt mà, giải phóng trọn vẹn tài nguyên và xử lý nốt các sự kiện còn đọng trong hàng đợi.
