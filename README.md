# Go Basic Roadmap & TaskManager System

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Dependencies](https://img.shields.io/badge/Dependencies-Zero%20External-success)](https://pkg.go.dev/std)
[![Tests](https://img.shields.io/badge/Tests-100%25%20Passing-brightgreen)](https://golang.org/cmd/go/#hdr-Test_packages)
[![Data Race](https://img.shields.io/badge/Race%20Detector-Clean-brightgreen)](https://go.dev/doc/articles/race_detector)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **Dự án mẫu chuẩn công nghiệp & Lộ trình học Go toàn diện**: Xây dựng hệ thống quản lý công việc (**TaskManager**) kết hợp cả **REST API Server** và **Command Line Interface (CLI)** thuần túy bằng Thư viện chuẩn của Go (**Zero External Dependencies**).

---

## Mục Lục (Table of Contents)

1. [Giới Thiệu Dự Án (Overview)](#giới-thiệu-dự-án-overview)
2. [Kiến Trúc Hệ Thống (Architecture)](#kiến-trúc-hệ-thống-architecture)
3. [Cấu Trúc Thư Mục & Quy Ước Package (Directory Layout)](#cấu-trúc-thư-mục--quy-ước-package-directory-layout)
4. [Yêu Cầu Môi Trường (Prerequisites)](#yêu-cầu-môi-trường-prerequisites)
5. [Khởi Động Nhanh (Quickstart)](#khởi-động-nhanh-quickstart)
6. [Cấu Hình Ứng Dụng (Configuration)](#cấu-hình-ứng-dụng-configuration)
7. [Tài Liệu REST API Đầy Đủ (REST API Reference)](#tài-liệu-rest-api-đầy-đủ-rest-api-reference)
8. [Tài Liệu Công Cụ Dòng Lệnh (CLI Reference)](#tài-liệu-công-cụ-dòng-lệnh-cli-reference)
9. [Chạy Kiểm Thử Tự Động (Running Tests & Quality Gates)](#chạy-kiểm-thử-tự-động-running-tests--quality-gates)
10. [Lộ Trình Học & Tài Liệu Đi Kèm (Roadmap & Documentation)](#lộ-trình-học--tài-liệu-đi-kèm-roadmap--documentation)

---

## Giới Thiệu Dự Án (Overview)

Dự án **TaskManager** được phát triển như một tài liệu tham khảo sống (Living Reference Implementation) dành cho lập trình viên muốn làm chủ Go từ nền tảng đến thực chiến:

- **100% Thư Viện Chuẩn (Zero External Dependencies)**: Không sử dụng Gin, Echo, Chi, GORM, hay bất kỳ thư viện bên ngoài nào. Tận dụng tối đa sức mạnh của Go 1.22+ (`net/http.ServeMux` hỗ trợ method & path parameter routing, `log/slog` structured logging, `sync.RWMutex`, `sync.WaitGroup`, `crypto/rand`, `os/signal`).
- **Kiến Trúc Phân Lớp Sạch (Clean Layered Architecture)**: Phân tách rõ ràng giữa Domain Models, Repository, Service, Worker Pool, Transport Handlers, và Middlewares.
- **Lập Trình Đồng Thời An Toàn (Safe Concurrency)**: Xử lý sự kiện ngầm bất đồng bộ bằng mô hình Worker Pool (Goroutines & Buffered Channels) kết hợp cơ chế dừng an toàn (Graceful Shutdown) và loại trừ Data Race.
- **Bộ Đôi Entrypoints**: Cung cấp đồng thời REST API Server phục vụ Web/Mobile clients và CLI Tool phục vụ tự động hóa qua terminal.

---

## Kiến Trúc Hệ Thống (Architecture)

```text
┌────────────────────────────────────────────────────────────────────────┐
│                        DUAL ENTRYPOINTS                                │
│        cmd/server/main.go                       cmd/cli/main.go        │
└───────────────────┬────────────────────────────────┬───────────────────┘
                    │                                │
                    ▼                                │
┌──────────────────────────────────────┐             │
│            HTTP TRANSPORT            │             │
│   internal/handler/task_handler.go   │             │
│   internal/middleware/middleware.go  │             │
│  (Go 1.22+ net/http ServeMux routing)│             │
└───────────────────┬──────────────────┘             │
                    │                                │
                    ▼                                ▼
┌────────────────────────────────────────────────────────────────────────┐
│                         CORE SERVICE LAYER                             │
│                  internal/service/task_service.go                      │
│        (Business Rules, Validation, Event Dispatching)                 │
└───────────────────┬────────────────────────────────┬───────────────────┘
                    │                                │
                    ▼                                ▼
┌──────────────────────────────────────┐  ┌──────────────────────────────┐
│        STORAGE ABSTRACTION           │  │    ASYNC WORKER POOL         │
│   internal/repository/repository.go  │  │ internal/worker/worker_pool.go│
│  - MemoryRepository (sync.RWMutex)   │  │ (Goroutines, Buffered Chan,  │
│  - FileRepository (Atomic JSON)      │  │  Graceful Stop, Context)     │
└───────────────────┬──────────────────┘  └──────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│                         DOMAIN MODELS & ERRORS                         │
│   internal/model/task.go                  internal/model/errors.go     │
└────────────────────────────────────────────────────────────────────────┘
```

---

## Cấu Trúc Thư Mục & Quy Ước Package (Directory Layout)

Cấu trúc dự án tuân thủ nghiêm ngặt theo **Standard Go Project Layout**:

```text
go-basic-roadmap/
├── cmd/
│   ├── server/             # Binary entrypoint cho REST API Server
│   │   └── main.go
│   └── cli/                # Binary entrypoint cho Command Line Interface Tool
│       └── main.go
├── internal/               # Mã nguồn nội bộ (Go ngăn chặn import từ bên ngoài)
│   ├── config/             # Tải cấu hình từ biến môi trường
│   │   └── config.go
│   ├── model/              # Domain entities, enums, validation, custom errors
│   │   ├── task.go
│   │   ├── task_test.go
│   │   └── errors.go
│   ├── repository/         # Data persistence interface & implementations
│   │   ├── repository.go   # TaskRepository interface contract
│   │   ├── memory.go       # In-memory thread-safe storage (sync.RWMutex)
│   │   ├── memory_test.go
│   │   ├── file.go         # Atomic JSON file-backed persistence
│   │   └── file_test.go
│   ├── service/            # Tầng nghiệp vụ điều phối logic & event
│   │   ├── task_service.go
│   │   └── task_service_test.go
│   ├── worker/             # Asynchronous concurrency worker pool
│   │   ├── worker_pool.go
│   │   └── worker_pool_test.go
│   ├── handler/            # HTTP handlers & JSON request/response helpers
│   │   ├── task_handler.go
│   │   ├── task_handler_test.go
│   │   └── response.go
│   └── middleware/         # HTTP Middlewares (Logging, Recovery, CORS)
│       └── middleware.go
├── docs/
│   ├── roadmap.md          # Lộ trình 6 giai đoạn học Go toàn diện từ số 0
│   └── step-by-step-guide.md # Hướng dẫn thực hành từng bước xây dựng dự án
├── Makefile                # Công cụ tự động hóa biên dịch, chạy và kiểm thử
├── go.mod                  # Định danh Go Module (Go 1.22+)
└── README.md               # Tài liệu tổng quan & hướng dẫn sử dụng
```

### Quy Ước Đặt Tên Package trong Go
1. **Ngắn gọn, rõ nghĩa**: Tên package viết bằng chữ thường (`model`, `service`, `worker`), không dùng `snake_case` hay `camelCase`.
2. **Tránh tên chung chung**: Không đặt tên package là `util`, `common`, hay `helpers`. Mỗi package có một trách nhiệm duy nhất (Single Responsibility).
3. **Bảo vệ bằng `internal/`**: Bất kỳ package nào nằm trong thư mục `internal/` đều được trình biên dịch Go bảo vệ, các module bên ngoài không thể import trái phép.

---

## Yêu Cầu Môi Trường (Prerequisites)

- **Go**: Phiên bản `1.22` trở lên (kiểm tra bằng `go version`).
- **Make**: Tiện ích chạy lệnh tự động (có sẵn trên macOS/Linux, Windows có thể dùng MinGW/WSL hoặc chạy trực tiếp lệnh Go).
- **cURL**: Tiện ích gọi HTTP API qua dòng lệnh.

Kiểm tra phiên bản Go trên máy:

```bash
go version
# Kết quả kỳ vọng: go version go1.22.x darwin/arm64 (hoặc linux/amd64, windows/amd64)
```

---

## Khởi Động Nhanh (Quickstart)

### 1. Biên Dịch Dự Án (Build)

Sử dụng `Makefile` để biên dịch cả Server và CLI:

```bash
make build
```

Hai tệp thực thi sẽ được xuất ra thư mục `bin/`:
- `bin/server`
- `bin/task-cli`

*(Hoặc sử dụng lệnh Go thuần túy: `go build -o bin/server ./cmd/server && go build -o bin/task-cli ./cmd/cli`)*

### 2. Khởi Chạy REST API Server

Chạy server mặc định (lưu trữ in-memory, cổng `:8080`):

```bash
make run-server
```

Server sẽ hiển thị log dạng JSON:

```json
{"time":"2026-11-02T02:30:00Z","level":"INFO","msg":"Starting TaskManager server","port":":8080","storage_type":"memory","worker_pool_size":3}
{"time":"2026-11-02T02:30:00Z","level":"INFO","msg":"Initialized MemoryTaskRepository"}
{"time":"2026-11-02T02:30:00Z","level":"INFO","msg":"Starting worker pool","workers":3,"buffer_capacity":100}
{"time":"2026-11-02T02:30:00Z","level":"INFO","msg":"HTTP server listening on :8080"}
```

### 3. Kiểm Tra Sức Khỏe Máy Chủ (Health Check)

Mở tab terminal mới và thực thi:

```bash
curl -s http://localhost:8080/health
```

Kết quả:
```json
{"status":"healthy","timestamp":"2026-11-02T02:30:05Z"}
```

### 4. Sử Dụng Công Cụ Dòng Lệnh (CLI)

Chạy CLI để thêm và xem công việc:

```bash
# Thêm công việc mới vào tệp JSON
./bin/task-cli add -title "Tìm hiểu Go Concurrency" -desc "Đọc Worker Pool" -priority HIGH

# Liệt kê danh sách công việc dạng bảng
./bin/task-cli list
```

---

## Cấu Hình Ứng Dụng (Configuration)

TaskManager hỗ trợ cấu hình linh hoạt thông qua các biến môi trường:

| Tên Biến Môi Trường | Kiểu Dữ Liệu | Mặc Định | Mô Tả Chi Tiết |
| :--- | :--- | :--- | :--- |
| `SERVER_PORT` / `PORT` | `string` | `:8080` | Cổng mạng TCP mà HTTP server lắng nghe (ví dụ: `:3000` hoặc `8080`). |
| `STORAGE_TYPE` | `string` | `memory` | Loại bộ lưu trữ: `memory` (lưu RAM) hoặc `file` (lưu tệp tin JSON bền vững). |
| `DATA_FILE_PATH` | `string` | `data/tasks.json` | Đường dẫn tệp tin JSON khi `STORAGE_TYPE=file`. |
| `WORKER_POOL_SIZE` | `int` | `3` | Số lượng goroutines worker chạy nền xử lý sự kiện bất đồng bộ. |
| `WORKER_QUEUE_SIZE` | `int` | `100` | Sức chứa tối đa của hàng đợi buffered channel nhận sự kiện. |
| `LOG_LEVEL` | `string` | `info` | Mức độ chi tiết của log: `debug`, `info`, `warn`, `error`. |

### Ví Dụ Chạy Với Lưu Trữ Tệp JSON & Cổng Tùy Chỉnh

```bash
STORAGE_TYPE=file DATA_FILE_PATH=data/my_tasks.json SERVER_PORT=:9000 ./bin/server
```

---

## Tài Liệu REST API Đầy Đủ (REST API Reference)

Base URL: `http://localhost:8080`

### Bảng Tổng Hợp Endpoint

| Phương thức | Đường dẫn | Chức năng | Mã trạng thái thành công |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | Kiểm tra trạng thái máy chủ | `200 OK` |
| `POST` | `/api/v1/tasks` | Tạo mới một công việc | `201 Created` |
| `GET` | `/api/v1/tasks` | Liệt kê danh sách công việc (có lọc) | `200 OK` |
| `GET` | `/api/v1/tasks/{id}` | Lấy chi tiết công việc theo ID | `200 OK` |
| `PUT` | `/api/v1/tasks/{id}` | Cập nhật thông tin công việc | `200 OK` |
| `DELETE` | `/api/v1/tasks/{id}` | Xóa công việc theo ID | `204 No Content` |

---

### Chi Tiết Từng Endpoint & Lệnh cURL Thực Tế

#### 1. Kiểm Tra Sức Khỏe Hệ Thống (Health Check)
- **URL**: `GET /health`
- **cURL Request**:
  ```bash
  curl -i http://localhost:8080/health
  ```
- **Response Header**: `HTTP/1.1 200 OK`, `Content-Type: application/json; charset=utf-8`
- **Response Body**:
  ```json
  {
    "status": "healthy",
    "timestamp": "2026-11-02T02:35:00Z"
  }
  ```

---

#### 2. Tạo Mới Công Việc (Create Task)
- **URL**: `POST /api/v1/tasks`
- **Headers**: `Content-Type: application/json`
- **Request Body Schema**:
  | Trường | Kiểu | Bắt buộc | Ràng buộc |
  | :--- | :--- | :--- | :--- |
  | `title` | `string` | **Có** | Không rỗng, tối đa 100 ký tự |
  | `description` | `string` | Không | Nội dung mô tả chi tiết |
  | `priority` | `string` | Không | `LOW`, `MEDIUM` (mặc định), hoặc `HIGH` |

- **cURL Request**:
  ```bash
  curl -i -X POST http://localhost:8080/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d '{
      "title": "Hoàn thành tài liệu Roadmap Go",
      "description": "Viết đầy đủ README và step-by-step guide",
      "priority": "HIGH"
    }'
  ```
- **Response Header**: `HTTP/1.1 201 Created`
- **Response Body**:
  ```json
  {
    "id": "e8a1f2b4-7b1e-4c5a-9351-826d03cf1e90",
    "title": "Hoàn thành tài liệu Roadmap Go",
    "description": "Viết đầy đủ README và step-by-step guide",
    "status": "TODO",
    "priority": "HIGH",
    "created_at": "2026-11-02T02:35:10.123456Z",
    "updated_at": "2026-11-02T02:35:10.123456Z"
  }
  ```

---

#### 3. Liệt Kê Danh Sách Công Việc (List Tasks)
- **URL**: `GET /api/v1/tasks`
- **Query Parameters**:
  - `status`: Lọc theo trạng thái (`TODO`, `IN_PROGRESS`, `DONE`).
  - `priority`: Lọc theo độ ưu tiên (`LOW`, `MEDIUM`, `HIGH`).
  - `search`: Tìm kiếm từ khóa không phân biệt hoa thường trong `title` hoặc `description`.

- **cURL Request (Lấy toàn bộ)**:
  ```bash
  curl -i http://localhost:8080/api/v1/tasks
  ```
- **cURL Request (Lọc theo điều kiện kết hợp)**:
  ```bash
  curl -i "http://localhost:8080/api/v1/tasks?status=TODO&priority=HIGH&search=Roadmap"
  ```
- **Response Header**: `HTTP/1.1 200 OK`
- **Response Body**:
  ```json
  [
    {
      "id": "e8a1f2b4-7b1e-4c5a-9351-826d03cf1e90",
      "title": "Hoàn thành tài liệu Roadmap Go",
      "description": "Viết đầy đủ README và step-by-step guide",
      "status": "TODO",
      "priority": "HIGH",
      "created_at": "2026-11-02T02:35:10.123456Z",
      "updated_at": "2026-11-02T02:35:10.123456Z"
    }
  ]
  ```

---

#### 4. Xem Chi Tiết Công Việc Theo ID (Get Task by ID)
- **URL**: `GET /api/v1/tasks/{id}`
- **cURL Request**:
  ```bash
  curl -i http://localhost:8080/api/v1/tasks/e8a1f2b4-7b1e-4c5a-9351-826d03cf1e90
  ```
- **Response Header**: `HTTP/1.1 200 OK`
- **Response Body**:
  ```json
  {
    "id": "e8a1f2b4-7b1e-4c5a-9351-826d03cf1e90",
    "title": "Hoàn thành tài liệu Roadmap Go",
    "description": "Viết đầy đủ README và step-by-step guide",
    "status": "TODO",
    "priority": "HIGH",
    "created_at": "2026-11-02T02:35:10.123456Z",
    "updated_at": "2026-11-02T02:35:10.123456Z"
  }
  ```
- **Trường Hợp Không Tìm Thấy (HTTP 404)**:
  ```bash
  curl -i http://localhost:8080/api/v1/tasks/non-existent-id
  ```
  ```json
  HTTP/1.1 404 Not Found
  Content-Type: application/json; charset=utf-8

  {
    "error": "task not found",
    "code": 404
  }
  ```

---

#### 5. Cập Nhật Công Việc (Update Task)
- **URL**: `PUT /api/v1/tasks/{id}`
- **Headers**: `Content-Type: application/json`
- **Request Body Schema**:
  | Trường | Kiểu | Bắt buộc | Ràng buộc |
  | :--- | :--- | :--- | :--- |
  | `title` | `string` | **Có** | Không rỗng, tối đa 100 ký tự |
  | `description` | `string` | Không | Mô tả cập nhật |
  | `status` | `string` | **Có** | `TODO`, `IN_PROGRESS`, hoặc `DONE` |
  | `priority` | `string` | **Có** | `LOW`, `MEDIUM`, hoặc `HIGH` |

- **cURL Request**:
  ```bash
  curl -i -X PUT http://localhost:8080/api/v1/tasks/e8a1f2b4-7b1e-4c5a-9351-826d03cf1e90 \
    -H "Content-Type: application/json" \
    -d '{
      "title": "Hoàn thành tài liệu Roadmap Go",
      "description": "Đang chạy kiểm thử nghiệm thu toàn bộ các tầng",
      "status": "IN_PROGRESS",
      "priority": "HIGH"
    }'
  ```
- **Response Header**: `HTTP/1.1 200 OK`
- **Response Body**:
  ```json
  {
    "id": "e8a1f2b4-7b1e-4c5a-9351-826d03cf1e90",
    "title": "Hoàn thành tài liệu Roadmap Go",
    "description": "Đang chạy kiểm thử nghiệm thu toàn bộ các tầng",
    "status": "IN_PROGRESS",
    "priority": "HIGH",
    "created_at": "2026-11-02T02:35:10.123456Z",
    "updated_at": "2026-11-02T02:36:00.654321Z"
  }
  ```

---

#### 6. Xóa Công Việc (Delete Task)
- **URL**: `DELETE /api/v1/tasks/{id}`
- **cURL Request**:
  ```bash
  curl -i -X DELETE http://localhost:8080/api/v1/tasks/e8a1f2b4-7b1e-4c5a-9351-826d03cf1e90
  ```
- **Response Header**: `HTTP/1.1 204 No Content`
- *(Không có nội dung thân phản hồi)*

---

## Tài Liệu Công Cụ Dòng Lệnh (CLI Reference)

Cú pháp tổng quát:
```bash
./bin/task-cli [flags] <command> [arguments]
```

Cờ toàn cục:
- `-data <đường_dẫn_file>`: Tùy chỉnh tệp tin JSON lưu trữ (Mặc định: `data/tasks.json`).

### 1. `add`: Thêm Công Việc Mới

Các cờ hỗ trợ:
- `-title <chuỗi>`: Tiêu đề công việc (**Bắt buộc**).
- `-desc <chuỗi>`: Mô tả công việc (Tùy chọn).
- `-priority <mức>`: Độ ưu tiên: `LOW`, `MEDIUM` (mặc định), `HIGH`.

```bash
./bin/task-cli add -title "Đọc tài liệu Net/HTTP" -desc "Tìm hiểu ServeMux Go 1.22" -priority HIGH
```
*Kết quả:*
```text
Task created successfully!
ID:          38b29fa0-482a-4364-a095-207d4b0f9dc8
Title:       Đọc tài liệu Net/HTTP
Description: Tìm hiểu ServeMux Go 1.22
Status:      TODO
Priority:    HIGH
Created:     Mon, 02 Nov 2026 09:40:00 +07
Updated:     Mon, 02 Nov 2026 09:40:00 +07
```

### 2. `list`: Liệt Kê Danh Sách Công Việc

Các cờ lọc:
- `-status <TODO|IN_PROGRESS|DONE>`: Lọc theo trạng thái.
- `-priority <LOW|MEDIUM|HIGH>`: Lọc theo độ ưu tiên.
- `-search <từ_khóa>`: Tìm kiếm trong tiêu đề hoặc mô tả.

```bash
./bin/task-cli list
```
*Kết quả hiển thị dạng bảng chuẩn:*
```text
ID                                     STATUS   PRIORITY   TITLE                    CREATED
--                                     ------   --------   -----                    -------
38b29fa0-482a-4364-a095-207d4b0f9dc8   TODO     HIGH       Đọc tài liệu Net/HTTP    2026-11-02 09:40
```

### 3. `get`: Xem Chi Tiết Công Việc
```bash
./bin/task-cli get 38b29fa0-482a-4364-a095-207d4b0f9dc8
```

### 4. `update`: Cập Nhật Công Việc

Cú pháp: `./bin/task-cli update <ID> [flags]`
- `-title`: Đổi tiêu đề.
- `-desc`: Đổi mô tả.
- `-status`: Đổi trạng thái (`TODO`, `IN_PROGRESS`, `DONE`).
- `-priority`: Đổi độ ưu tiên (`LOW`, `MEDIUM`, `HIGH`).

```bash
./bin/task-cli update 38b29fa0-482a-4364-a095-207d4b0f9dc8 -status IN_PROGRESS
```

### 5. `delete`: Xóa Công Việc
```bash
./bin/task-cli delete 38b29fa0-482a-4364-a095-207d4b0f9dc8
```
*Kết quả:*
```text
Task "38b29fa0-482a-4364-a095-207d4b0f9dc8" deleted successfully.
```

---

## Chạy Kiểm Thử Tự Động (Running Tests & Quality Gates)

Dự án sở hữu bộ kiểm thử tự động toàn diện bao phủ toàn bộ các tầng logic:

### 1. Chạy Tất Cả Unit Tests
```bash
make test
# Tương đương: go test -v ./...
```

### 2. Kiểm Tra Xung Đột Bộ Nhớ Đa Luồng (Race Detector)
Trình phát hiện race condition của Go sẽ giám sát các truy cập bộ nhớ đồng thời:
```bash
make test-race
# Tương đương: go test -v -race ./...
```

### 3. Phân Tích Độ Bao Phủ Mã Nguồn (Code Coverage)
Đo lường phần trăm dòng lệnh đã được kiểm thử:
```bash
make test-coverage
# Tương đương: go test -v -race -coverprofile=coverage.out -covermode=atomic ./... && go tool cover -func=coverage.out
```

### 4. Định Dạng & Kiểm Tra Tĩnh (Format & Vet)
```bash
make fmt  # Định dạng code chuẩn gofmt
make vet  # Phân tích tĩnh phát hiện lỗi logic tiềm ẩn bằng go vet
```

---

## Lộ Trình Học & Tài Liệu Đi Kèm (Roadmap & Documentation)

Dự án này là một phần trong hệ thống học liệu Go chuẩn công nghiệp:

1. 📖 **[Lộ Trình Học Go Toàn Diện 6 Giai Đoạn (docs/roadmap.md)](docs/roadmap.md)**:
   - Giai đoạn 1: Thiết lập môi trường & Hệ sinh thái công cụ Go (`go mod`, `gopls`, Delve).
   - Giai đoạn 2: Nền tảng ngôn ngữ & Cú pháp đặc trưng (Variables, Control flow, Slices, Maps, Pointers).
   - Giai đoạn 3: Hướng đối tượng kiểu Go (Structs, Methods, Value vs Pointer Receivers, Implicit Interfaces).
   - Giai đoạn 4: Lập trình đồng thời đỉnh cao (Goroutines, Channels, CSP, Mutex, Deadlock, Race Conditions).
   - Giai đoạn 5: Cấu trúc dự án chuẩn & Quy ước thiết kế Package.
   - Giai đoạn 6: Phân tích kiến trúc TaskManager Reference Project.
   - Bảng thuật ngữ chuyên ngành Anh - Việt đầy đủ.

2. 🛠️ **[Hướng Dẫn Thực Hành Xây Dựng Từng Bước (docs/step-by-step-guide.md)](docs/step-by-step-guide.md)**:
   - Bản hướng dẫn thực hành chi tiết từng bước (Step-by-step Hands-on Tutorial) giúp bạn viết từng dòng mã nguồn từ con số 0 đến khi hoàn thiện toàn bộ hệ thống.

---

## Tác Giả & Giấy Phép (Author & License)

- **Mục tiêu giáo dục**: Cung cấp kiến thức lập trình Go vững chắc, thực tế, nói không với phụ thuộc thư viện mù quáng.
- **Giấy phép**: Phát hành theo giấy phép [MIT License](LICENSE). Tự do học tập, sử dụng và chia sẻ vì cộng đồng Go Việt Nam.
