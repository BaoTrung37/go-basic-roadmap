# Lộ Trình Học Go (Golang) Toàn Diện: Từ Số 0 Đến Dự Án Thực Chiến

> **Mục tiêu**: Hướng dẫn người học làm chủ ngôn ngữ lập trình Go (Golang) theo phương pháp tiếp cận bậc thang 6 giai đoạn (6-stage scaffolding). Tài liệu kết hợp lý thuyết bản chất, phân tích mô hình bộ nhớ, cảnh báo các cạm bẫy kinh điển (common pitfalls), bài tập thực hành có lời giải chi tiết và hướng dẫn xây dựng dự án thực tế chuẩn công nghiệp (**TaskManager**) với tiêu chí **Zero External Dependencies** (chỉ sử dụng Thư viện chuẩn Go 1.22+).

---

## Mục Lục Tổng Quan

1. [Triết Lý Cốt Lõi Của Go (The Go Philosophy)](#triết-lý-cốt-lõi-của-go-the-go-philosophy)
2. [Giai Đoạn 1: Thiết Lập Môi Trường & Hệ Sinh Thái Công Cụ (Environment & Toolchain Setup)](#giai-đoạn-1-thiết-lập-môi-trường--hệ-sinh-thái-công-cụ)
3. [Giai Đoạn 2: Nền Tảng Ngôn Ngữ & Cú Pháp Đặc Trưng (Core Go Fundamentals)](#giai-đoạn-2-nền-tảng-ngôn-ngữ--cú-pháp-đặc-trưng)
4. [Giai Đoạn 3: Lập Trình Hướng Đối Tượng Kiểu Go (Go-Style OOP: Structs, Methods & Interfaces)](#giai-đoạn-3-lập-trình-hướng-đối-tượng-kiểu-go)
5. [Giai Đoạn 4: Lập Trình Đồng Thời (Concurrency with CSP: Goroutines & Channels)](#giai-đoạn-4-lập-trình-đồng-thời)
6. [Giai Đoạn 5: Cấu Trúc Dự Án Chuẩn & Quy Ước Đặt Tên (Standard Project Layout & Conventions)](#giai-đoạn-5-cấu-trúc-dự-án-chuẩn--quy-ước-đặt-tên)
7. [Giai Đoạn 6: Hướng Dẫn Xây Dựng Dự Án Thực Tế Từng Bước (Practical Project Walkthrough: TaskManager)](#giai-đoạn-6-hướng-dẫn-xây-dựng-dự-án-thực-tế-từng-bước)
8. [Tài Nguyên Học Tập Tiếp Theo (Next Steps & Advanced Topics)](#tài-nguyên-học-tập-tiếp-theo)

---

## Triết Lý Cốt Lõi Của Go (The Go Philosophy)

Trước khi viết dòng code đầu tiên, người học cần thấu hiểu tư duy thiết kế đã định hình nên Go:

1. **Đơn giản hơn là phức tạp (Simplicity over Complexity)**: Go chủ ý loại bỏ những tính năng rườm rà như kế thừa phân cấp (`class inheritance`), hàm quá tải (`function overloading`), ném ngoại lệ (`exceptions/try-catch`), toán tử con trỏ (`pointer arithmetic`). Sự tối giản này giúp mã nguồn Go luôn dễ đọc, dễ duy trì và đồng nhất giữa các lập trình viên.
2. **Tường minh hơn là ngầm định (Explicit over Implicit)**: Mọi luồng điều khiển, chuyển đổi kiểu dữ liệu, và xử lý lỗi đều được thể hiện rõ ràng trên từng dòng mã. Không có phép chuyển kiểu ngầm định (no implicit type casting), không có luồng thực thi ẩn giấu (no hidden control flow).
3. **Thành phần hóa thay vì Kế thừa (Composition over Inheritance)**: Thay vì xây dựng các cây kế thừa đồ sộ và giòn gãy (fragile base class problem), Go kết hợp các `struct` nhỏ và `interface` rời rạc thông qua cơ chế *Struct Embedding* và *Implicit Interfaces*.
4. **Lập trình đồng thời dựa trên CSP (Communicating Sequential Processes)**: Thay vì chia sẻ vùng nhớ chung và dùng khóa mutex phức tạp, Go đề xướng phương châm:
   > *"Do not communicate by sharing memory; instead, share memory by communicating."*  
   *(Đừng giao tiếp bằng cách chia sẻ bộ nhớ; hãy chia sẻ bộ nhớ bằng cách giao tiếp thông qua Channel).*
5. **Lỗi là giá trị thông thường (Errors are values)**: Trong Go, lỗi không phải là ngoại lệ làm ngắt luồng chương trình. Lỗi là một giá trị trả về thông thường, buộc lập trình viên phải đối mặt và xử lý ngay tại vị trí phát sinh.

---

## GIAI ĐOẠN 1: Thiết Lập Môi Trường & Hệ Sinh Thái Công Cụ

### 1.1. Mục Tiêu Học Tập (Learning Outcomes)
- Cài đặt thành công Go Toolchain phiên bản mới nhất (Go 1.22+ khuyến nghị) trên hệ điều hành.
- Nắm vững cấu trúc biến môi trường cốt lõi của Go (`GOROOT`, `GOPATH`, `GOBIN`, `GOPROXY`).
- Phân biệt bản chất lịch sử: Kỷ nguyên `$GOPATH mode` (trước Go 1.11) và Kỷ nguyên `Go Modules` (hiện đại, chuẩn hóa từ Go 1.16+).
- Cấu hình IDE chuyên nghiệp (VS Code hoặc GoLand) với Language Server (`gopls`), Delve debugger (`dlv`), và cơ chế tự động định dạng (`gofmt`, `goimports`).
- Làm chủ bộ công cụ dòng lệnh (CLI toolchain): `go mod init`, `go run`, `go build`, `go vet`, `go fmt`, `go test`.

---

### 1.2. Nền Tảng Lý Thuyết (Theoretical Background)

#### Cài Đặt và Kiểm Tra Biến Môi Trường
Truy cập [go.dev/dl](https://go.dev/dl/) để tải bản cài đặt tương ứng với hệ điều hành của bạn (macOS, Linux hoặc Windows). Sau khi cài đặt, mở terminal và kiểm tra:

```bash
# Kiểm tra phiên bản Go hiện tại
go version
# Kết quả mẫu: go version go1.22.5 darwin/arm64 hoặc go1.24+

# Xem toàn bộ cấu hình môi trường Go
go env
```

Các biến môi trường Go quan trọng nhất cần hiểu rõ:
- **`GOROOT`**: Thư mục cài đặt chính của Go SDK chứa trình biên dịch và toàn bộ mã nguồn của Thư viện chuẩn (`std`). Thường được hệ thống tự cấu hình (ví dụ: `/usr/local/go` hoặc `/opt/homebrew/opt/go/libexec`). Bạn **không** nên tự ý can thiệp vào thư mục này.
- **`GOPATH`**: Thư mục làm việc cá nhân của lập trình viên (mặc định là `~/go` trên macOS/Linux hoặc `%USERPROFILE%\go` trên Windows). Trong mô hình hiện đại, `GOPATH` chủ yếu dùng làm nơi lưu trữ bộ nhớ đệm của các package tải về (`$GOPATH/pkg/mod`) và các file thực thi được cài qua lệnh `go install` (`$GOPATH/bin`).
- **`GOBIN`**: Thư mục chứa các tệp thực thi sinh ra từ `go install`. Khuyến nghị thêm `$GOBIN` (hoặc `$GOPATH/bin`) vào biến `PATH` của hệ điều hành.
- **`GOPROXY`**: Proxy trung gian phục vụ việc tải dependencies an toàn và tốc độ cao (mặc định là `https://proxy.golang.org,direct`).

#### Lịch Sử: `$GOPATH Mode` vs `Go Modules Mode`

| Đặc điểm so sánh | Kỷ nguyên Cũ (`$GOPATH Mode`) | Kỷ nguyên Hiện Đại (`Go Modules`) |
| :--- | :--- | :--- |
| **Vị trí thư mục dự án** | Bắt buộc phải nằm trong `$GOPATH/src/github.com/user/project` | Đặt tại **bất kỳ thư mục nào** trên ổ cứng của bạn |
| **Quản lý phiên bản** | Không có cơ chế phiên bản chính thức (luôn lấy HEAD của branch master) | Khai báo phiên bản chính xác qua `go.mod` (Semantic Versioning `vX.Y.Z`) |
| **Bảo toàn tính toàn vẹn** | Dễ bị thay đổi mã nguồn ngầm định | Được bảo vệ tuyệt đối bằng hàm băm mật mã trong `go.sum` |
| **Cấu hình kích hoạt** | Cần set `GO111MODULE=off` (nay đã bị khai tử hoàn toàn) | Tự động kích hoạt khi có tệp `go.mod` trong thư mục gốc |

#### Cấu Hình Trình Soạn Thảo (Editor Tooling)
Go sở hữu hệ sinh thái công cụ hỗ trợ lập trình viên (Developer Experience - DX) thuộc hàng tốt nhất hiện nay:
- **`gopls` (Go Language Server)**: Cung cấp tính năng tự động hoàn thành mã (auto-complete), nhảy tới định nghĩa (`Go to Definition`), tìm kiếm tham chiếu (`Find References`), hiển thị gợi ý kiểu dữ liệu (Hover Information) và tái cấu trúc mã (Refactoring).
- **`gofmt` & `goimports`**: `gofmt` là công cụ định dạng code chính thức của Go, chấm dứt mọi tranh cãi về thụt lề, dấu ngoặc nhọn hay khoảng trắng. `goimports` là bản mở rộng của `gofmt`, tự động thêm các import còn thiếu và xóa các import thừa.
- **`dlv` (Delve)**: Trình gỡ lỗi (debugger) chuyên biệt dành cho Go, hỗ trợ bắt breakpoint, xem trạng thái goroutines, biến và stack trace.

#### Các Lệnh CLI Cốt Lõi
1. `go mod init <module-path>`: Khởi tạo module mới, tạo tệp `go.mod`.
2. `go run <file.go>`: Biên dịch tạm thời trong bộ nhớ và chạy ngay chương trình (thích hợp trong giai đoạn phát triển).
3. `go build -o <output> <package>`: Biên dịch mã nguồn thành file nhị phân độc lập (self-contained executable binary), không cần bất kỳ runtime hay máy ảo nào để thực thi.
4. `go vet ./...`: Công cụ phân tích tĩnh (Static Analysis) quét tìm các lỗi ngữ nghĩa tiềm ẩn mà trình biên dịch có thể bỏ qua (ví dụ: in sai tham số format `Printf`, khóa Mutex bị copy).
5. `go test -v ./...`: Chạy toàn bộ các bài kiểm thử tự động trong dự án.
6. `go mod tidy`: Tự động tìm nạp các package phụ thuộc còn thiếu và dọn dẹp các package thừa trong `go.mod` và `go.sum`.

---

### 1.3. Cạm Bẫy Thường Gặp (Common Pitfalls)

#### Cạm bẫy 1.1: Quên khởi tạo `go.mod` trước khi viết code
- **Hiện tượng**: Bạn tạo file `main.go`, tạo thêm file helper `math/calc.go`, sau đó import `calc` vào `main.go`. Khi chạy `go run main.go`, chương trình báo lỗi:
  ```text
  package math/calc is not in std (/usr/local/go/src/math/calc)
  ```
- **Nguyên nhân**: Không có tệp `go.mod`, Go không xác định được Module Path gốc của dự án và mặc định tìm kiếm package trong Thư viện chuẩn (`GOROOT`).
- **Giải pháp**: Luôn chạy `go mod init <tên-module>` ngay tại thư mục gốc trước khi bắt đầu dự án.

#### Cạm bẫy 1.2: Code không được format chuẩn và xuất hiện khoảng trắng rác
- **Hiện tượng**: Pull Request bị từ chối vì xung đột định dạng (tab vs space, dấu ngoặc xuống dòng).
- **Nguyên nhân**: Chưa bật tính năng "Format on Save" sử dụng `gofmt` hoặc `goimports` trong IDE. Trong Go, quy ước định dạng là quy chuẩn mang tính bắt buộc của cộng đồng: mở ngoặc nhọn `{` phải nằm cùng dòng với khai báo hàm/cấu trúc điều khiển, thụt đầu dòng bằng Tab thật (`\t`).

---

### 1.4. Bài Tập Thực Hành Giai Đoạn 1

#### Bài tập 1.1: Khởi tạo Module và Kiểm tra Môi trường Hệ thống
- **Yêu cầu**: Tạo một module có tên `sysinfo`, viết hàm `main` kiểm tra và in ra thông tin: Hệ điều hành (`runtime.GOOS`), Kiến trúc CPU (`runtime.GOARCH`), Số lượng CPU Cores khả dụng (`runtime.NumCPU()`), và phiên bản Go hiện hành (`runtime.Version()`).
- **Mã nguồn giải chi tiết**:

```go
// File: sysinfo/main.go
package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("=== THÔNG TIN HỆ THỐNG GO ===")
	fmt.Printf("Hệ điều hành (OS)      : %s\n", runtime.GOOS)
	fmt.Printf("Kiến trúc vi xử lý (Arch): %s\n", runtime.GOARCH)
	fmt.Printf("Số nhân CPU logic (Cores): %d\n", runtime.NumCPU())
	fmt.Printf("Phiên bản Go hiện hành  : %s\n", runtime.Version())
}
```

- **Các bước thực thi trên terminal**:
```bash
mkdir -p sysinfo && cd sysinfo
go mod init sysinfo
# Tạo file main.go với nội dung trên
go run main.go
go build -o bin/sysinfo main.go
./bin/sysinfo
```

#### Bài tập 1.2: Phân Tích Mã Bằng `go vet`
- **Yêu cầu**: Viết một đoạn mã cố tình vi phạm định dạng chuỗi trong `fmt.Printf` (ví dụ: dùng `%d` cho một chuỗi string). Quan sát phản hồi của `go build` so với phản hồi của `go vet`.
- **Mã nguồn thử nghiệm**:

```go
// File: vetdemo/main.go
package main

import "fmt"

func main() {
	name := "Gopher"
	// Lỗi: %d là định dạng số nguyên, nhưng name là chuỗi string
	fmt.Printf("Xin chào: %d\n", name)
}
```

- **Thực hành kiểm tra**:
```bash
# go build có thể vẫn biên dịch thành công (chỉ in ra cảnh báo runtime %!d(string=Gopher))
go build -o /dev/null main.go

# Chạy go vet để phát hiện lỗi ngay tại bước kiểm tra tĩnh:
go vet ./...
# Kết quả hiển thị chính xác lỗi:
# ./main.go:8:2: fmt.Printf format %d has arg name of wrong type string
```

---

## GIAI ĐOẠN 2: Nền Tảng Ngôn Ngữ & Cú Pháp Đặc Trưng

### 2.1. Mục Tiêu Học Tập (Learning Outcomes)
- Hiểu sâu cơ chế khai báo biến (`var`, `:=`), quy tắc Zero Values và Phạm vi truy cập (Scope).
- Phân biệt bản chất bộ nhớ giữa Value Types (Array, Struct cơ bản) và Reference/Descriptor Types (Slice, Map, Channel, Pointer).
- Nắm vững cấu trúc bộ nhớ của Slice Header (`Data pointer`, `Length`, `Capacity`) và cơ chế tái cấp phát bộ nhớ ngầm định khi gọi `append`.
- Làm chủ cấu trúc điều khiển: `if` kèm câu lệnh khởi tạo ngắn (short statement), vòng lặp `for` (từ khóa lặp duy nhất), và `switch` (không tự động fallthrough).
- Hiểu cơ chế hoạt động của hàm (First-class functions), đa giá trị trả về (Multiple return values), và cơ chế trì hoãn thực thi `defer` (nguyên tắc LIFO, bắt cặp tài nguyên).
- Nắm vững triết lý xử lý lỗi: "Errors are values", tạo lỗi định danh (sentinel errors), bọc lỗi ngữ cảnh (`%w`), và kiểm tra lỗi bằng `errors.Is` cùng `errors.As`.

---

### 2.2. Nền Tảng Lý Thuyết (Theoretical Background)

#### Khai Báo Biến & Zero Values
Go là ngôn ngữ định kiểu tĩnh an toàn (Statically Typed). Go loại bỏ khái niệm vùng nhớ rác (uninitialized garbage memory): biến sau khi khai báo luôn nhận giá trị mặc định (**Zero Value**):
- Số nguyên (`int`, `int64`, ...): `0`
- Số thực (`float32`, `float64`): `0.0`
- Boolean (`bool`): `false`
- Chuỗi ký tự (`string`): `""` (chuỗi rỗng, không phải `nil`)
- Con trỏ (`*T`), Slice (`[]T`), Map (`map[K]V`), Channel, Interface, Function: `nil`

Cú pháp khai báo:
```go
var age int = 25       // Khai báo tường minh đầy đủ
var name = "Alice"     // Type inference (Go tự suy luận kiểu string)
var uninitialized int  // Nhận zero-value là 0

// Khai báo ngắn (Short variable declaration) - Chỉ được dùng bên trong thân hàm:
score := 95.5
```

#### Arrays vs. Slices: Bản Chất Bộ Nhớ
- **Array (Mảng)**: Là tập hợp tuần tự các phần tử có **kích thước cố định ngay tại thời điểm biên dịch**. Độ dài là một phần của kiểu dữ liệu (`[3]int` là kiểu dữ liệu hoàn toàn khác với `[4]int`). Khi gán một array cho biến khác, Go sẽ **sao chép toàn bộ dữ liệu** (Pass-by-value).
- **Slice (Lát cắt)**: Là một cấu trúc mô tả (Descriptor / Slice Header) bao bọc lấy một underlying array bên dưới. Kích thước của Slice Header trên hệ thống 64-bit là chính xác 24 bytes, gồm 3 trường:
  1. `Data unsafe.Pointer`: Con trỏ trỏ tới phần tử đầu tiên của mảng bên dưới.
  2. `Len int`: Số lượng phần tử hiện có trong slice (8 bytes).
  3. `Cap int`: Dung lượng tối đa mà slice có thể chứa trước khi phải tái cấp phát bộ nhớ mới (8 bytes).

```text
Slice Header (24 bytes):
┌────────────────┬──────────┬──────────┐
│  Data Pointer  │ Len = 3  │ Cap = 5  │
└───────┬────────┴──────────┴──────────┘
        │
        ▼ Underlying Array trong Heap/Stack:
       [ 10 | 20 | 30 | (trống) | (trống) ]
         0    1    2       3         4
```

Khi bạn gọi `append(slice, val)`:
- Nếu `Len < Cap`: Go điền giá trị vào ô trống tiếp theo trong underlying array và tăng `Len` lên 1. Không có việc cấp phát bộ nhớ mới xảy ra.
- Nếu `Len == Cap`: Go sẽ cấp phát một mảng underlying array mới có kích thước gấp đôi (theo thuật toán phân bổ bộ nhớ của runtime), sao chép toàn bộ phần tử cũ sang, thêm phần tử mới, và trả về một Slice Header mới trỏ tới mảng mới này.

#### Bản Chất Của Map
Map trong Go là một con trỏ trỏ tới cấu trúc hash table nội bộ (`runtime.hmap`).
- Map **không an toàn cho việc đọc ghi đồng thời (thread-unsafe)**.
- Phải khởi tạo Map trước khi ghi dữ liệu bằng hàm `make(map[K]V)` hoặc map literal `{}`.

#### Cấu Trúc Điều Khiển Đặc Trưng
1. **`if` với Short Statement**: Cho phép khai báo biến tạm thời, phạm vi của biến chỉ tồn tại trong các khối `if` và `else`:
   ```go
   if val, ok := myMap[key]; ok {
       fmt.Println("Tìm thấy:", val)
   } else {
       fmt.Println("Không tìm thấy khóa:", key)
   }
   // val và ok không thể truy cập bên ngoài khối này!
   ```
2. **Vòng lặp `for` duy nhất**: Go loại bỏ hoàn toàn các từ khóa `while` hay `do-while`. Mọi dạng lặp đều viết bằng `for`:
   ```go
   // Dạng lặp 3 thành phần chuẩn:
   for i := 0; i < 5; i++ { ... }

   // Dạng lặp kiểu while:
   for condition { ... }

   // Dạng lặp vô tận:
   for { ... }

   // Dạng duyệt qua Slice/Map bằng range:
   for index, value := range mySlice { ... }
   for key, val := range myMap { ... }
   ```
3. **`switch` không tự động Fallthrough**: Trong Go, khi một `case` thỏa mãn, chương trình sẽ thoát ngay khỏi `switch` mà không nhảy xuống case tiếp theo (trừ khi cố tình dùng từ khóa `fallthrough`).

#### Cơ Chế Trì Hoãn `defer`
Lệnh `defer` đăng ký một hàm được thực thi ngay trước khi hàm chứa nó trả về kết quả (return).
- Các lệnh `defer` được quản lý theo cấu trúc ngăn xếp LIFO (Last-In, First-Out).
- Tham số truyền vào hàm `defer` được tính toán ngay tại dòng khai báo `defer`, nhưng thân hàm chỉ chạy khi hàm bao quanh kết thúc.
- Ứng dụng quan trọng nhất: Dọn dẹp tài nguyên (Resource Cleanup) — đóng file, giải phóng kết nối mạng, mở khóa Mutex.

```go
func processFile(filename string) error {
    f, err := os.Open(filename)
    if err != nil {
        return err
    }
    // Đảm bảo file luôn được đóng bất kể hàm return ở nhánh nào phía sau:
    defer f.Close()

    // Tiến hành đọc ghi file an toàn...
    return nil
}
```

#### Triết Lý Xử Lý Lỗi Hiện Đại (Errors are Values)
Trong Go:
1. Lỗi là một giá trị thỏa mãn interface `error`:
   ```go
   type error interface {
       Error() string
   }
   ```
2. Không sử dụng `panic/recover` cho các lỗi nghiệp vụ thông thường. Chỉ dùng `panic` khi hệ thống gặp lỗi bất khả kháng không thể tiếp tục chạy (ví dụ thiếu file cấu hình cốt lõi lúc khởi động).
3. **Error Wrapping (Từ Go 1.13+)**: Dùng động từ `%w` trong `fmt.Errorf` để đóng gói lỗi gốc kèm thêm ngữ cảnh:
   ```go
   if err != nil {
       return fmt.Errorf("không thể kết nối tới cơ sở dữ liệu: %w", err)
   }
   ```
4. Kiểm tra lỗi bằng `errors.Is` (kiểm tra lỗi cụ thể trong chuỗi bọc lỗi) và `errors.As` (ép kiểu trích xuất kiểu lỗi tùy biến):
   ```go
   if errors.Is(err, sql.ErrNoRows) {
       // Xử lý trường hợp không tìm thấy dữ liệu
   }
   ```

---

### 2.3. Cạm Bẫy Thường Gặp (Common Pitfalls)

#### Cạm bẫy 2.1: Gán dữ liệu vào Nil Map gây Panic
- **Hiện tượng**:
  ```go
  var userAges map[string]int // Zero value là nil
  userAges["Alice"] = 28      // PANIC: assignment to entry in nil map
  ```
- **Giải pháp**: Luôn khởi tạo map bằng `make` hoặc literal:
  ```go
  userAges := make(map[string]int)
  userAges["Alice"] = 28 // Hoàn toàn an toàn
  ```

#### Cạm bẫy 2.2: Hiểu sai cơ chế sao chép giá trị trong `for ... range`
- **Hiện tượng**: Thay đổi biến lặp trong `for range` không làm thay đổi giá trị trong slice gốc:
  ```go
  numbers := []int{1, 2, 3}
  for _, val := range numbers {
      val = val * 2 // VÔ NGHĨA! val chỉ là bản sao tạm thời của phần tử
  }
  // numbers vẫn là [1, 2, 3]
  ```
- **Giải pháp**: Truy cập và gán trực tiếp qua chỉ mục:
  ```go
  for i := range numbers {
      numbers[i] = numbers[i] * 2
  }
  // numbers trở thành [2, 4, 6]
  ```

#### Cạm bẫy 2.3: Đặt `defer` bên trong vòng lặp lớn gây cạn kiệt tài nguyên
- **Hiện tượng**:
  ```go
  for _, filename := range millionsOfFiles {
      f, _ := os.Open(filename)
      defer f.Close() // NGUY HIỂM: defer chỉ chạy khi toàn bộ hàm kết thúc!
  }
  ```
- **Giải pháp**: Tách thân vòng lặp thành một hàm độc lập (anonymous function hoặc helper function) để `defer` thực thi ngay sau mỗi lần lặp.

---

### 2.4. Bài Tập Thực Hành Giai Đoạn 2

#### Bài tập 2.1: Bộ Đếm Tần Suất Từ Vựng (Word Frequency Counter)
- **Yêu cầu**: Viết hàm `CountWordFrequency(text string) (map[string]int, error)`. Hàm nhận vào một chuỗi văn bản, phân tách thành các từ (bỏ qua khoảng trắng thừa, chuẩn hóa chữ thường), đếm tần suất xuất hiện của từng từ. Nếu chuỗi đầu vào rỗng, trả về một lỗi định danh tùy biến `ErrEmptyText`.
- **Mã nguồn giải chi tiết**:

```go
// File: wordcount/wordcount.go
package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var ErrEmptyText = errors.New("văn bản đầu vào không được để trống")

func CountWordFrequency(text string) (map[string]int, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, ErrEmptyText
	}

	// Tách từ dựa trên các ký tự phân cách (khoảng trắng, dấu câu)
	words := strings.FieldsFunc(trimmed, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})

	frequency := make(map[string]int, len(words))
	for _, word := range words {
		cleaned := strings.ToLower(strings.TrimSpace(word))
		if cleaned != "" {
			frequency[cleaned]++
		}
	}

	return frequency, nil
}

func main() {
	doc := "Go là ngôn ngữ đơn giản. Go rất nhanh, và Go rất mạnh mẽ!"
	res, err := CountWordFrequency(doc)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	fmt.Println("Kết quả đếm tần suất từ vựng:")
	for word, count := range res {
		fmt.Printf("- Từ '%s': %d lần\n", word, count)
	}

	// Thử nghiệm nhánh lỗi:
	_, errEmpty := CountWordFrequency("   ")
	if errors.Is(errEmpty, ErrEmptyText) {
		fmt.Println("\nKiểm tra lỗi thành công: Bắt được ErrEmptyText chính xác.")
	}
}
```

#### Bài tập 2.2: Sao Chép File An Toàn Kèm Bọc Lỗi Đa Tầng
- **Yêu cầu**: Viết hàm `SafeCopyFile(src, dst string) error` sao chép nội dung từ file nguồn tới file đích. Yêu cầu:
  1. Kiểm tra tồn tại file nguồn.
  2. Dùng `defer` để đóng file nguồn và file đích đúng quy chuẩn.
  3. Bọc lỗi với `%w` để bên ngoài có thể truy vết ngữ cảnh.
- **Mã nguồn giải chi tiết**:

```go
// File: filecopy/filecopy.go
package main

import (
	"fmt"
	"io"
	"os"
)

func SafeCopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("không thể mở file nguồn '%s': %w", src, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("không thể tạo file đích '%s': %w", dst, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("lỗi trong quá trình sao chép từ '%s' sang '%s': %w", src, dst, err)
	}

	// Đồng bộ dữ liệu xuống đĩa vật lý trước khi kết thúc
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("lỗi đồng bộ dữ liệu đĩa cho '%s': %w", dst, err)
	}

	return nil
}

func main() {
	src := "source_test.txt"
	dst := "dest_test.txt"

	// Tạo file nguồn tạm thời
	_ = os.WriteFile(src, []byte("Dữ liệu thử nghiệm sao chép an toàn trong Go\n"), 0644)
	defer os.Remove(src)
	defer os.Remove(dst)

	if err := SafeCopyFile(src, dst); err != nil {
		fmt.Printf("Sao chép thất bại: %v\n", err)
		return
	}

	fmt.Println("Sao chép file an toàn thành công!")
	content, _ := os.ReadFile(dst)
	fmt.Printf("Nội dung file đích: %s", string(content))
}
```

---

## GIAI ĐOẠN 3: Lập Trình Hướng Đối Tượng Kiểu Go

### 3.1. Mục Tiêu Học Tập (Learning Outcomes)
- Nắm vững cú pháp định nghĩa `struct`, anonymous struct và struct tags (ví dụ `json:"id"`).
- Phân biệt rõ Value Semantics vs Pointer Semantics: Khi nào dùng Value Receiver (`func (t Task)`) và khi nào bắt buộc dùng Pointer Receiver (`func (t *Task)`).
- Hiểu sâu cơ chế Interfaces của Go: Implicit Implementation (Duck Typing tĩnh tại compile-time, không cần từ khóa `implements`), triết lý Small Interfaces (chỉ chứa 1-2 phương thức).
- Nắm rõ cấu trúc của Interface Value gồm cặp `(type, value)` và cạm bẫy chết người `typed nil interface`.
- Áp dụng triệt để nguyên lý **Composition over Inheritance** qua kỹ thuật Struct Embedding.
- Xây dựng mô hình Dependency Injection (DI) thuần khiết (Pure DI) mà không cần dùng bất kỳ framework phản chiếu (reflection) phức tạp nào.

---

### 3.2. Nền Tảng Lý Thuyết (Theoretical Background)

#### Struct & Struct Tags
Struct là kiểu dữ liệu tổng hợp (composite type) đóng gói các trường dữ liệu có liên quan:
```go
type Task struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
}
```
Struct tags (chuỗi nằm trong dấu backtick \` `) cung cấp metadata cho các thư viện xử lý thông qua reflection (như `encoding/json` chuyển đổi giữa Go struct và JSON payload).

#### Value Receiver vs. Pointer Receiver
Phương thức (Method) trong Go là một hàm gắn liền với một kiểu dữ liệu cụ thể thông qua tham số nhận (Receiver Parameter):

| Tiêu chí so sánh | Value Receiver: `func (t Task) Read()` | Pointer Receiver: `func (t *Task) Update()` |
| :--- | :--- | :--- |
| **Hành vi bộ nhớ** | Tạo ra một **bản sao độc lập** của struct tại thời điểm gọi | Truyền địa chỉ con trỏ, chia sẻ vùng nhớ gốc |
| **Khả năng sửa đổi** | Bất kỳ thay đổi nào trên trường chỉ ảnh hưởng lên bản sao, struct gốc không đổi | **Sửa đổi trực tiếp** trạng thái của struct gốc |
| **Hiệu năng (Performance)** | Tốn chi phí sao chép nếu struct lớn (nhiều trường dữ liệu) | Rất nhanh (chỉ truyền con trỏ 8 bytes) |
| **Tính đồng bộ (`sync.Mutex`)** | **Tuyệt đối cấm** nếu struct có chứa `sync.Mutex` (Mutex bị copy sẽ mất tác dụng khóa) | Bắt buộc sử dụng |

> **Nguyên tắc vàng (Rule of Thumb)**: Nếu một phương thức của struct cần thay đổi trạng thái hoặc struct đó chứa Mutex, hãy dùng Pointer Receiver cho **toàn bộ** các phương thức của struct đó để đảm bảo tính nhất quán (Consistency).

#### Implicit Interfaces: Trực Giác Của "Duck Typing Tĩnh"
Trong Java/C#, bạn phải dùng từ khóa `class MyRepo implements Repository`.  
Trong Go, nếu một struct sở hữu đầy đủ danh sách các phương thức được khai báo trong interface, struct đó **tự động thỏa mãn interface đó tại thời điểm biên dịch**. Không có từ khóa `implements`.

```go
// Định nghĩa interface tối giản
type Greeter interface {
    Greet() string
}

type VietnameseSpeaker struct{}

// VietnameseSpeaker tự động thỏa mãn Greeter mà không cần khai báo tường minh
func (v VietnameseSpeaker) Greet() string {
    return "Xin chào các bạn!"
}
```

Lợi ích to lớn: Tầng nghiệp vụ (Service layer) có thể tự định nghĩa interface cho dữ liệu nó cần, và tầng lưu trữ (Repository layer) cung cấp cài đặt mà hai tầng không bị kết dính chặt chẽ (Decoupled).

#### Cấu Trúc Interface Value & Cạm Bẫy "Typed Nil"
Một biến Interface trong Go chiếm 16 bytes gồm hai con trỏ bên dưới:
1. `type descriptor`: Trỏ tới kiểu dữ liệu cụ thể (concrete type).
2. `value pointer`: Trỏ tới giá trị dữ liệu cụ thể.

Một biến interface chỉ thực sự bằng `nil` khi **cả hai con trỏ `type == nil` VÀ `value == nil`**.

```text
Interface bình thường:
┌─────────────────────┬─────────────────────┐
│  Type = *MyError    │  Value = 0x1234F    │  => interface != nil (Hợp lệ)
└─────────────────────┴─────────────────────┘

Interface Nil thuần túy:
┌─────────────────────┬─────────────────────┐
│  Type = nil         │  Value = nil        │  => interface == nil (Hợp lệ)
└─────────────────────┴─────────────────────┘

Cạm bẫy Typed Nil:
┌─────────────────────┬─────────────────────┐
│  Type = *MyError    │  Value = nil        │  => interface != nil (Rất nguy hiểm!)
└─────────────────────┴─────────────────────┘
```

#### Composition Over Inheritance (Struct Embedding)
Go không có từ khóa `extends`. Để tái sử dụng mã nguồn, Go nhúng struct con vào struct cha. Các trường và phương thức của struct con sẽ được tự động thăng hạng (promoted):

```go
type BaseModel struct {
    ID        string
    CreatedAt time.Time
}

type User struct {
    BaseModel // Struct Embedding (Anonymous field)
    Username  string
}

// Khi sử dụng:
u := User{}
u.ID = "usr_001" // Truy cập thẳng trường ID của BaseModel được thăng hạng!
```

---

### 3.3. Cạm Bẫy Thường Gặp (Common Pitfalls)

#### Cạm bẫy 3.1: Dùng Value Receiver làm mất thay đổi trạng thái
- **Hiện tượng**:
  ```go
  type Counter struct{ Count int }
  func (c Counter) Increment() { c.Count++ } // Value receiver!

  c := Counter{Count: 0}
  c.Increment()
  // c.Count vẫn là 0! Vì c bên trong Increment chỉ là một bản sao chép.
  ```
- **Giải pháp**: Đổi thành `func (c *Counter) Increment() { c.Count++ }`.

#### Cạm bẫy 3.2: Lỗi Typed Nil Interface khi trả về Custom Error
- **Hiện tượng**:
  ```go
  type CustomError struct{}
  func (e *CustomError) Error() string { return "lỗi tùy biến" }

  func check() error {
      var err *CustomError = nil // Con trỏ cụ thể bằng nil
      if someCondition {
          err = &CustomError{}
      }
      return err // NGUY HIỂM: Trả về interface có Type=*CustomError và Value=nil
  }

  func main() {
      if check() != nil {
          // Khối này LUÔN ĐƯỢC CHẠY ngay cả khi err cụ thể là nil!
          fmt.Println("Có lỗi xảy ra!") 
      }
  }
  ```
- **Giải pháp**: Luôn trả về giá trị `nil` tường minh cho kiểu interface `error`:
  ```go
  func check() error {
      if someCondition {
          return &CustomError{}
      }
      return nil // Trả về nil thuần túy (Type=nil, Value=nil)
  }
  ```

---

### 3.4. Bài Tập Thực Hành Giai Đoạn 3

#### Bài tập 3.1: Struct Model & Pointer Receiver
- **Yêu cầu**: Xây dựng struct `Task` có các trường `ID`, `Title`, `Completed`, `Priority` (`LOW`, `MEDIUM`, `HIGH`). Viết các phương thức sử dụng Pointer Receiver: `MarkCompleted()`, `SetPriority(p Priority) error`.
- **Mã nguồn giải chi tiết**:

```go
// File: taskmodel/task.go
package main

import (
	"fmt"
	"time"
)

type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityMedium Priority = "MEDIUM"
	PriorityHigh   Priority = "HIGH"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	Priority  Priority  `json:"priority"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *Task) MarkCompleted() {
	t.Completed = true
	t.UpdatedAt = time.Now()
}

func (t *Task) SetPriority(p Priority) error {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		t.Priority = p
		t.UpdatedAt = time.Now()
		return nil
	default:
		return fmt.Errorf("độ ưu tiên '%s' không hợp lệ", p)
	}
}

func main() {
	task := &Task{
		ID:        "tsk_101",
		Title:     "Học Go Object-Oriented Programming",
		Completed: false,
		Priority:  PriorityMedium,
		UpdatedAt: time.Now(),
	}

	task.MarkCompleted()
	if err := task.SetPriority(PriorityHigh); err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	fmt.Printf("Task ID: %s | Title: %s | Completed: %t | Priority: %s\n",
		task.ID, task.Title, task.Completed, task.Priority)
}
```

#### Bài tập 3.2: Thiết Kế Interface & In-Memory Repository Với Dependency Injection
- **Yêu cầu**: Định nghĩa interface `TaskRepository` với các phương thức `Save(task *Task) error` và `FindByID(id string) (*Task, error)`. Triển khai struct `MemoryTaskRepository` sử dụng map an toàn, sau đó inject vào `TaskService`.
- **Mã nguồn giải chi tiết**:

```go
// File: di_repo/main.go
package main

import (
	"errors"
	"fmt"
	"sync"
)

var ErrTaskNotFound = errors.New("không tìm thấy task với ID chỉ định")

// 0. Định nghĩa Domain Model
type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityMedium Priority = "MEDIUM"
	PriorityHigh   Priority = "HIGH"
)

type Task struct {
	ID        string
	Title     string
	Completed bool
	Priority  Priority
}

// 1. Khai báo Interface
type TaskRepository interface {
	Save(task *Task) error
	FindByID(id string) (*Task, error)
}

// 2. Cài đặt Struct cụ thể (In-Memory Repository)
type MemoryTaskRepository struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewMemoryTaskRepository() *MemoryTaskRepository {
	return &MemoryTaskRepository{
		tasks: make(map[string]*Task),
	}
}

func (r *MemoryTaskRepository) Save(task *Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *MemoryTaskRepository) FindByID(id string) (*Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, exists := r.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// 3. Tầng Service nhận Repository qua Constructor (Pure Dependency Injection)
type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(id, title string) (*Task, error) {
	task := &Task{
		ID:        id,
		Title:     title,
		Completed: false,
		Priority:  PriorityMedium,
	}
	if err := s.repo.Save(task); err != nil {
		return nil, fmt.Errorf("không thể lưu task: %w", err)
	}
	return task, nil
}

func main() {
	// Khởi tạo các phụ thuộc và ghép nối (Wiring)
	repo := NewMemoryTaskRepository()
	service := NewTaskService(repo)

	created, err := service.CreateTask("t1", "Làm chủ Go Interfaces")
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}
	fmt.Printf("Đã tạo task thành công: %+v\n", created)

	found, err := repo.FindByID("t1")
	if err != nil {
		fmt.Println("Lỗi tìm kiếm:", err)
		return
	}
	fmt.Printf("Tìm thấy task trong Repository: ID=%s, Title=%s\n", found.ID, found.Title)
}
```

---

## GIAI ĐOẠN 4: Lập Trình Đồng Thời

### 4.1. Mục Tiêu Học Tập (Learning Outcomes)
- Thấu hiểu mô hình đồng thời CSP (Communicating Sequential Processes) và cơ chế hoạt động của Go Runtime Scheduler (M:N Scheduler).
- Nắm vững vòng đời của Goroutines: Khởi tạo siêu nhẹ (~2KB stack ban đầu), tự động co giãn kích thước stack theo nhu cầu.
- Làm chủ Channels: Unbuffered Channels (đồng bộ hóa tức thì / rendezvous) vs Buffered Channels (hàng đợi bất đồng bộ có kích thước), Directional Channels (`chan<-` và `<-chan`).
- Nắm vững 4 quy tắc sống còn của Channel Lifecycle để tránh panic và deadlock.
- Sử dụng thành thạo câu lệnh `select` để multiplexing đa kênh và cài đặt mẫu timeout với `time.After`.
- Sử dụng chuẩn xác gói `sync`: `sync.WaitGroup` để đồng bộ hoàn thành, `sync.Mutex` & `sync.RWMutex` bảo vệ vùng tranh chấp tài nguyên (Critical Section).
- Quản lý vòng đời request và hủy bỏ tác vụ đồng thời bằng gói `context` (`context.WithCancel`, `context.WithTimeout`).
- Phát hiện và dập tắt hoàn toàn các lỗi nguy hiểm: Goroutine Leaks, Data Race (`go test -race`), Deadlock.

---

### 4.2. Nền Tảng Lý Thuyết (Theoretical Background)

#### Goroutine vs. Hệ Điều Hành Thread (OS Thread)
- **OS Thread**: Do hạt nhân hệ điều hành (Kernel) quản lý. Mỗi thread tiêu tốn dung lượng bộ nhớ cố định từ 1MB đến 8MB cho stack, chi phí chuyển đổi ngữ cảnh (Context Switch) tốn hàng nghìn chu kỳ CPU do phải chuyển qua kernel space.
- **Goroutine**: Do Go Runtime tự quản lý hoàn toàn ở không gian người dùng (User space). Ban đầu chỉ tiêu tốn khoảng **2KB** bộ nhớ stack. Go Scheduler sử dụng mô hình lập lịch **M:N**: Ghép $N$ Goroutines lên $M$ OS Threads thông qua $P$ Logical Processors (`GOMAXPROCS`). Một tiến trình Go có thể chạy đồng thời hàng trăm nghìn Goroutines mà không làm kiệt quệ tài nguyên máy chủ.

#### Kênh Truyền Thông (Channels)
Channel là đường ống giao tiếp an toàn đa luồng kết nối giữa các goroutine:

1. **Unbuffered Channel (`make(chan T)`)**:
   - Dung lượng buffer = 0.
   - Thao tác gửi (`ch <- data`) sẽ **bị block** cho đến khi có một goroutine khác sẵn sàng nhận (`<-ch`), và ngược lại. Đây là cơ chế đồng bộ hóa bắt tay (Rendezvous Point).
2. **Buffered Channel (`make(chan T, capacity)`)**:
   - Có hàng đợi chứa tối đa `capacity` phần tử.
   - Thao tác gửi chỉ bị block khi buffer đã **đầy**.
   - Thao tác nhận chỉ bị block khi buffer đang **rỗng**.
3. **Directional Channels (Kênh có định hướng)**:
   - `chan<- T`: Kênh chỉ cho phép gửi dữ liệu vào.
   - `<-chan T`: Kênh chỉ cho phép nhận dữ liệu ra.
   - Giúp compiler kiểm tra tính đúng đắn của logic chương trình.

#### 4 Quy Tắc Sống Còn Khi Đóng Channel
1. **Chỉ Goroutine gửi (Sender) mới được quyền đóng channel**: Không bao giờ đóng channel từ phía nhận (Receiver).
2. **Gửi dữ liệu vào channel đã đóng (`closed channel`) sẽ gây PANIC ngay lập tức**.
3. **Đóng một channel đã đóng lần thứ hai sẽ gây PANIC ngay lập tức**.
4. **Nhận dữ liệu từ channel đã đóng vẫn an toàn tuyệt đối**: Trả về các phần tử còn sót lại trong buffer; khi buffer hết, trả về zero-value và cờ `ok == false`:
   ```go
   val, ok := <-ch
   if !ok {
       fmt.Println("Channel đã đóng và không còn dữ liệu!")
   }
   ```

#### Multiplexing Với `select`
Câu lệnh `select` tương tự như `switch`, nhưng mỗi case là một thao tác trên channel. `select` sẽ block cho đến khi một trong các case sẵn sàng giao tiếp:

```go
select {
case msg := <-ch1:
    fmt.Println("Nhận từ ch1:", msg)
case ch2 <- "dữ liệu":
    fmt.Println("Đã gửi thành công vào ch2")
case <-time.After(2 * time.Second):
    fmt.Println("Hết thời gian chờ (Timeout)!")
default:
    fmt.Println("Không có channel nào sẵn sàng (Non-blocking operation)")
}
```

#### Quản Lý Luồng Bằng `sync` & `context`
- **`sync.WaitGroup`**: Bộ đếm đồng bộ. Dùng `wg.Add(n)` để báo số tác vụ, `wg.Done()` trong goroutine khi hoàn thành, và `wg.Wait()` ở luồng chính để chờ tất cả hoàn tất.
- **`sync.RWMutex`**: Cho phép nhiều luồng cùng đọc đồng thời (`RLock()`), nhưng chỉ cho phép duy nhất một luồng ghi độc quyền (`Lock()`).
- **`context.Context`**: Cung cấp cơ chế truyền tín hiệu dừng xuyên suốt qua các tầng ứng dụng (HTTP Handler -> Service -> Database / Worker). Khi timeout hoặc người dùng ngắt kết nối, `ctx.Done()` phát tín hiệu đóng channel, giải phóng lập tức mọi goroutine đang chờ.

---

### 4.3. Cạm Bẫy Thường Gặp (Common Pitfalls)

#### Cạm bẫy 4.1: Rò rỉ Goroutine (Goroutine Leak)
- **Hiện tượng**: Khởi chạy goroutine lắng nghe trên một unbuffered channel hoặc gửi vào một unbuffered channel mà không có đối tác tương ứng:
  ```go
  func queryFastestServer() string {
      ch := make(chan string) // Unbuffered!
      go func() { ch <- fetch("server_a") }()
      go func() { ch <- fetch("server_b") }()
      return <-ch // Nhận kết quả từ server nhanh nhất rồi thoát hàm!
      // CẠM BẪY: Goroutine thứ hai sẽ bị block vĩnh viễn trên ch <- fetch!
      // Nó không bao giờ được Garbage Collector dọn dẹp -> Rò rỉ bộ nhớ.
  }
  ```
- **Giải pháp**: Sử dụng buffered channel với dung lượng đủ cho tất cả người gửi: `ch := make(chan string, 2)`.

#### Cạm bẫy 4.2: Truyền `sync.WaitGroup` qua hàm bằng giá trị
- **Hiện tượng**: Khai báo `func worker(wg sync.WaitGroup)` mà không dùng con trỏ `*sync.WaitGroup`. Go sẽ sao chép cấu trúc `WaitGroup`, lệnh `wg.Done()` chỉ giảm bộ đếm của bản sao, hàm `main` gọi `wg.Wait()` sẽ bị treo vĩnh viễn dẫn đến **Deadlock**.
- **Giải pháp**: Luôn truyền `wg *sync.WaitGroup`.

#### Cạm bẫy 4.3: Xung Đột Truy Cập Dữ Liệu (Data Race)
- **Hiện tượng**: Hai goroutine cùng đọc và ghi vào một biến hoặc map mà không có Mutex bảo vệ. Lỗi này có thể không gây crash ngay nhưng sẽ làm hỏng dữ liệu một cách ngẫu nhiên.
- **Giải pháp**: Luôn kiểm tra mã nguồn bằng cờ `-race`:
  ```bash
  go test -race ./...
  go run -race main.go
  ```

---

### 4.4. Bài Tập Thực Hành Giai Đoạn 4

#### Bài tập 4.1: Xây Dựng Worker Pool Chuẩn Mực
- **Yêu cầu**: Xây dựng một hệ thống xử lý công việc đồng thời (**Worker Pool**) với:
  - Một `Job` struct chứa ID và dữ liệu số nguyên.
  - Một `Result` struct chứa ID của Job, kết quả bình phương, và worker ID xử lý.
  - Khởi chạy cố định $N$ worker goroutines (ví dụ 3 workers).
  - Sử dụng `sync.WaitGroup` để điều phối đóng kết quả an toàn.
- **Mã nguồn giải chi tiết**:

```go
// File: workerpool/main.go
package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID    int
	Value int
}

type Result struct {
	Job       Job
	Square    int
	WorkerID  int
	Duration  time.Duration
}

// Hàm thực thi của từng Worker
func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		start := time.Now()
		// Giả lập thời gian tính toán
		time.Sleep(50 * time.Millisecond)
		results <- Result{
			Job:      job,
			Square:   job.Value * job.Value,
			WorkerID: id,
			Duration: time.Since(start),
		}
	}
}

func main() {
	const numJobs = 10
	const numWorkers = 3

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	var wg sync.WaitGroup

	// 1. Khởi chạy Worker Pool
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// 2. Đẩy các công việc vào jobs channel
	for j := 1; j <= numJobs; j++ {
		jobs <- Job{ID: j, Value: j * 2}
	}
	close(jobs) // Đóng kênh jobs để báo cho các worker biết đã hết việc

	// 3. Goroutine giám sát: Đợi tất cả worker xong thì đóng results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// 4. Thu thập và in kết quả từ results channel
	fmt.Println("=== KẾT QUẢ TỪ WORKER POOL ===")
	for res := range results {
		fmt.Printf("Job %2d (Input: %2d) -> Output: %4d | Xử lý bởi Worker %d (%v)\n",
			res.Job.ID, res.Job.Value, res.Square, res.WorkerID, res.Duration)
	}
	fmt.Println("Toàn bộ công việc đã được hoàn thành trọn vẹn!")
}
```

#### Bài tập 4.2: Gọi Tác Vụ Giả Lập Kèm Quản Lý Timeout Bằng Context
- **Yêu cầu**: Viết hàm `fetchDataWithTimeout(ctx context.Context, duration time.Duration) (string, error)`. Dùng `context.WithTimeout` để giới hạn thời gian chờ tối đa là 100ms. Thử nghiệm với tác vụ tốn 50ms (thành công) và tác vụ tốn 200ms (bị hủy bởi context).
- **Mã nguồn giải chi tiết**:

```go
// File: contextdemo/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func simulateRemoteCall(ctx context.Context, workDuration time.Duration) (string, error) {
	resChan := make(chan string, 1)

	go func() {
		time.Sleep(workDuration)
		resChan <- "Dữ liệu máy chủ trả về thành công!"
	}()

	select {
	case <-ctx.Done():
		// Bị ngắt do timeout hoặc cancel từ context cha
		return "", ctx.Err()
	case res := <-resChan:
		return res, nil
	}
}

func main() {
	// Kịch bản 1: Hoàn thành trước khi timeout (Success)
	ctx1, cancel1 := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel1()

	fmt.Println("1. Thử nghiệm tác vụ nhanh (50ms) với timeout 150ms:")
	data, err := simulateRemoteCall(ctx1, 50*time.Millisecond)
	if err != nil {
		fmt.Println("Thất bại:", err)
	} else {
		fmt.Println("Thành công:", data)
	}

	// Kịch bản 2: Vượt quá thời hạn (Deadline Exceeded)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()

	fmt.Println("\n2. Thử nghiệm tác vụ chậm (250ms) với timeout 100ms:")
	data2, err2 := simulateRemoteCall(ctx2, 250*time.Millisecond)
	if err2 != nil {
		if errors.Is(err2, context.DeadlineExceeded) {
			fmt.Println("Thành công bắt lỗi Timeout: Tác vụ bị hủy bởi context.DeadlineExceeded!")
		} else {
			fmt.Println("Lỗi khác:", err2)
		}
	} else {
		fmt.Println("Thành công:", data2)
	}
}
```

---

## GIAI ĐOẠN 5: Cấu Trúc Dự Án Chuẩn & Quy Ước Đặt Tên

### 5.1. Mục Tiêu Học Tập (Learning Outcomes)
- Nắm vững kiến trúc dự án Go chuẩn công nghiệp (**Standard Go Project Layout**).
- Thấu hiểu cơ chế bảo mật mã nguồn độc nhất vô nhị của thư mục `internal/` được **Go Compiler kiểm tra và cưỡng chế nghiêm ngặt**.
- Nắm bắt quy ước đặt tên Package theo phong cách Go chuẩn (Go Idiomatic Naming Conventions).
- Xóa bỏ triệt để các phản hoa văn thiết kế (Anti-patterns) tai hại như các package gom góp: `util`, `common`, `helpers`.
- Hiểu nguyên nhân sâu xa của lỗi nhập khẩu vòng lặp (Circular Dependency / Import Cycle) và làm chủ kỹ thuật đảo ngược phụ thuộc (Dependency Inversion) để triệt tiêu hoàn toàn lỗi này.
- Quản lý phiên bản dependencies tự động với `go.mod` và `go.sum`.

---

### 5.2. Nền Tảng Lý Thuyết (Theoretical Background)

#### Cấu Trúc Standard Go Project Layout

```text
my-go-service/
├── cmd/                        # Các điểm nhập (Entrypoints) chính của ứng dụng
│   ├── server/                 # Tệp nhị phân REST API server
│   │   └── main.go             # Khởi tạo config, wiring dependencies, lắng nghe shutdown
│   └── cli/                    # (Tùy chọn) Tệp nhị phân CLI thao tác quản trị
│       └── main.go
├── internal/                   # MÃ NGUỒN RIÊNG TƯ (Compiler-Enforced Private Packages)
│   ├── config/                 # Nạp biến môi trường và cấu hình hệ thống
│   ├── model/                  # Domain Entities, Structs, Enums, Custom Error Types
│   ├── repository/             # Giao diện lưu trữ (Interfaces) và các bản cài đặt (Memory, File, SQL)
│   ├── service/                # Logic nghiệp vụ trung tâm (Business Logic)
│   ├── worker/                 # Hệ thống Background Workers xử lý đồng thời
│   ├── handler/                # Tầng giao vận HTTP: nhận request, parse JSON, gọi service, trả HTTP status
│   └── middleware/             # Middlewares: ghi log có cấu trúc (slog), thu hồi panic
├── pkg/                        # (Tùy chọn) Các thư viện công cộng mở cho dự án bên ngoài import
├── docs/                       # Tài liệu hướng dẫn kiến trúc, lộ trình, API Swagger/OpenAPI
├── Makefile                    # Các lệnh tự động hóa: make run, make build, make test, make lint
├── go.mod                      # Định nghĩa Module path và phiên bản Go
└── go.sum                      # Checksum mật mã đảm bảo tính toàn vẹn của mã nguồn
```

#### Sức Mạnh Của Thư Mục `internal/`
Trong Go, thư mục mang tên `internal/` sở hữu đặc quyền được trình biên dịch Go bảo vệ ở cấp độ ngôn ngữ:
- Bất kỳ package nào nằm dưới cây thư mục `internal/` **chỉ có thể được import** bởi các package là anh em hoặc con cháu của thư mục cha trực tiếp chứa `internal`.
- Dự án khác (bên ngoài module) nếu cố tình import `my-module/internal/...` sẽ bị Go compiler chặn đứng ngay lập tức với lỗi:
  ```text
  use of internal package not allowed
  ```
- Điều này cho phép bạn thoải mái tái cấu trúc mã nguồn nội bộ mà không lo sợ làm hỏng mã nguồn của các bên phụ thuộc bên ngoài.

#### Quy Ước Đặt Tên Package (Package Naming Rules)
1. **Ngắn gọn, một từ duy nhất, viết thường hoàn toàn**:
   - Tốt: `package model`, `package handler`, `package task`.
   - Cấm: `package task_model`, `package taskModel`, `package task-service`.
2. **Tránh trùng lặp từ ngữ gây khó chịu (Avoid Stuttering)**:
   Khi gọi một hàm hay struct từ package khác, tên package luôn đứng đầu (ví dụ `task.Item`). Nếu bạn đặt struct là `TaskItem` trong package `task`, người dùng sẽ phải gõ `task.TaskItem` (bị lặp từ). Hãy đặt là `task.Item`.
3. **Loại bỏ vĩnh viễn các package "rác" (`util`, `common`, `helpers`)**:
   - Lý do: Những package này không thể hiện rõ trách nhiệm nghiệp vụ (Single Responsibility Principle). Chúng thu hút mọi hàm tiện ích linh tinh và là nguyên nhân số 1 dẫn đến lỗi **Circular Dependency**.
   - Khắc phục: Phân rã chúng theo đúng chức năng thực tế: `package hash`, `package validator`, `package crypto`.

#### Triệt Tiêu Lỗi Phụ Thuộc Vòng (Circular Dependency)
Go nghiêm cấm việc Package A import Package B và Package B lại import ngược lại Package A. Trình biên dịch sẽ báo lỗi ngay: `import cycle not allowed`.

**Giải pháp kiến trúc**: Luồng phụ thuộc một chiều từ ngoài vào trong:

```text
[cmd/server] -> [handler] -> [service] -> [repository] -> [model]
```
- Package `model` nằm ở đáy cùng, không phụ thuộc vào bất kỳ package nào trong `internal/`.
- Tầng `service` không phụ thuộc vào cài đặt cụ thể của `repository`, mà chỉ phụ thuộc vào `interface TaskRepository`.
- Việc ghép nối (wiring) bản cài đặt cụ thể `MemoryRepository` vào `Service` được thực hiện tại tầng ngoài cùng: `cmd/server/main.go`.

---

### 5.3. Cạm Bẫy Thường Gặp (Common Pitfalls)

#### Cạm bẫy 5.1: Nhồi nhét toàn bộ logic vào `cmd/server/main.go`
- **Hiện tượng**: File `main.go` phình to lên hàng ngàn dòng, chứa từ khai báo struct database đến routing HTTP.
- **Giải pháp**: Tuân thủ nguyên tắc "Thin Main". Hàm `main()` chỉ làm 4 nhiệm vụ tinh gọn:
  1. Đọc file cấu hình (`config.Load()`).
  2. Khởi tạo cơ sở dữ liệu / repository (`repo := repository.NewMemory()`).
  3. Khởi tạo service và handler (`svc := service.New(repo)`, `h := handler.New(svc)`).
  4. Lắng nghe tín hiệu hệ điều hành để tắt ứng dụng an toàn (Graceful Shutdown).

---

### 5.4. Bài Tập Thực Hành Giai Đoạn 5

#### Bài tập 5.1: Xây Dựng Khung Dự Án Chuẩn & Khắc Phục Circular Import
- **Yêu cầu**: Tạo cấu trúc module `taskflow` với hai package `internal/model` và `internal/service`. Chứng minh luồng dữ liệu 1 chiều không phát sinh lỗi circular import.
- **Cấu trúc thư mục**:

```bash
mkdir -p taskflow/cmd/server taskflow/internal/model taskflow/internal/service
cd taskflow
go mod init taskflow
```

- **Mã nguồn `internal/model/task.go`**:
```go
package model

import "time"

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}
```

- **Mã nguồn `internal/service/task_service.go`**:
```go
package service

import (
	"taskflow/internal/model"
	"time"
)

type TaskService struct{}

func NewTaskService() *TaskService {
	return &TaskService{}
}

func (s *TaskService) BuildSampleTask(id, title string) model.Task {
	return model.Task{
		ID:        id,
		Title:     title,
		CreatedAt: time.Now(),
	}
}
```

- **Mã nguồn `cmd/server/main.go`**:
```go
package main

import (
	"fmt"
	"taskflow/internal/service"
)

func main() {
	svc := service.NewTaskService()
	task := svc.BuildSampleTask("t-01", "Cấu trúc Go chuẩn")
	fmt.Printf("Khởi tạo thành công: ID=%s, Title=%s, CreatedAt=%v\n",
		task.ID, task.Title, task.CreatedAt)
}
```

- **Kiểm tra biên dịch**:
```bash
go run cmd/server/main.go
# Biên dịch và chạy hoàn hảo, không có phụ thuộc vòng lặp!
```

---

## GIAI ĐOẠN 6: Hướng Dẫn Xây Dựng Dự Án Thực Tế Từng Bước

> **Dự Án Mẫu Minh Họa**: **TaskManager REST API & CLI System**  
> Dự án áp dụng trọn vẹn toàn bộ kiến thức từ Giai đoạn 1 đến Giai đoạn 5, không sử dụng bất kỳ thư viện bên thứ ba nào (Zero External Dependencies - chỉ dùng Thư viện chuẩn Go 1.22+).

---

### 6.1. Kiến Trúc Tổng Thể Dự Án TaskManager

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
│   internal/middleware/logging.go     │             │
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

### 6.2. Lộ Trình 7 Bước Tự Tay Phát Triển Dự Án

#### Bước 1: Khởi Tạo Dự Án & Bộ Khung Thư Mục
Mở terminal và thực thi các lệnh sau để khởi tạo:
```bash
mkdir -p go-basic-roadmap && cd go-basic-roadmap
go mod init go-basic-roadmap

mkdir -p cmd/server cmd/cli
mkdir -p internal/config internal/model internal/repository internal/service internal/worker internal/handler internal/middleware
mkdir -p docs
```

#### Bước 2: Thiết Kế Domain Models & Phân Loại Lỗi (`internal/model`)
- Khai báo các trạng thái `Status` (`TODO`, `IN_PROGRESS`, `DONE`) và độ ưu tiên `Priority` (`LOW`, `MEDIUM`, `HIGH`).
- Tạo cấu trúc `Task` với đầy đủ JSON struct tags.
- Khai báo các lỗi nghiệp vụ định danh (`ErrTaskNotFound`, `ErrEmptyTitle`, `ErrInvalidStatus`, `ErrInvalidPriority`).
- Viết phương thức `Validate()` kiểm tra tính toàn vẹn dữ liệu trước khi lưu trữ.

```go
// Tóm tắt cấu trúc Task trong internal/model/task.go:
type Task struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      Status    `json:"status"`
    Priority    Priority  `json:"priority"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### Bước 3: Thiết Kế Repository Interface & Bộ Lưu Trữ Đa Hình (`internal/repository`)
- Định nghĩa interface `TaskRepository` với chuẩn chữ ký phương thức:
  ```go
  type TaskRepository interface {
      Create(ctx context.Context, task *model.Task) error
      GetByID(ctx context.Context, id string) (*model.Task, error)
      List(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error)
      Update(ctx context.Context, task *model.Task) error
      Delete(ctx context.Context, id string) error
  }
  ```
- Cài đặt `MemoryRepository`: Dùng `map[string]*model.Task` kết hợp `sync.RWMutex` bảo vệ đa luồng.
- Cài đặt `FileRepository`: Ghi dữ liệu ra tệp JSON vật lý an toàn (Atomic File Write: ghi vào file tạm rồi dùng `os.Rename` thay thế nguyên tử để không bao giờ bị corrupt file khi mất điện đột ngột).

#### Bước 4: Xây Dựng Tầng Nghiệp Vụ & Worker Pool Xử Lý Sự Kiện Bất Đồng Bộ
- **`internal/worker/worker_pool.go`**:
  - Nhận các sự kiện `model.TaskEvent` (khi task được tạo, sửa, xóa) qua một Buffered Channel.
  - Phân phối cho các worker goroutines xử lý ghi log thống kê ngầm.
  - Hỗ trợ dừng an toàn (`Stop()`) đảm bảo xử lý hết hàng đợi trước khi ngắt ứng dụng.
- **`internal/service/task_service.go`**:
  - Tiếp nhận input từ HTTP hoặc CLI.
  - Kiểm tra validation nghiệp vụ.
  - Gọi Repository để lưu dữ liệu.
  - Phát sự kiện sang Worker Pool bất đồng bộ mà không làm chậm thời gian phản hồi HTTP request.

#### Bước 5: Xây Dựng HTTP Transport Layer Sử Dụng Go 1.22+ (`internal/handler`)
Từ Go 1.22+, `net/http.ServeMux` chính thức hỗ trợ định tuyến theo phương thức HTTP và trích xuất tham số trên URL mà không cần cài đặt thêm thư viện ngoài như Gorilla Mux hay Chi:

```go
mux := http.NewServeMux()

// Cú pháp định tuyến chuẩn Go 1.22+:
mux.HandleFunc("POST /api/v1/tasks", h.CreateTask)
mux.HandleFunc("GET /api/v1/tasks", h.ListTasks)
mux.HandleFunc("GET /api/v1/tasks/{id}", h.GetTask)
mux.HandleFunc("PUT /api/v1/tasks/{id}", h.UpdateTask)
mux.HandleFunc("DELETE /api/v1/tasks/{id}", h.DeleteTask)

// Trích xuất path value cực kỳ đơn giản:
id := r.PathValue("id")
```

Kết hợp các Middlewares chuyên dụng tại `internal/middleware`:
- **Logging Middleware**: Ghi log mọi request (Method, Path, RemoteAddr, Latency, Status Code) bằng thư viện chuẩn `log/slog`.
- **Recovery Middleware**: Dùng `defer` và `recover()` bắt gọn mọi panic phát sinh trong handler, tránh làm sập máy chủ và phản hồi lỗi `500 Internal Server Error` về cho client.

#### Bước 6: Khởi Tạo Ứng Dụng & Graceful Shutdown (`cmd/server/main.go`)
- Đọc cấu hình từ biến môi trường (Port, Storage Type).
- Khởi tạo các tầng theo thứ tự: `WorkerPool` -> `Repository` -> `Service` -> `Handler`.
- Sử dụng `signal.NotifyContext` lắng nghe các tín hiệu hệ điều hành (`SIGINT`, `SIGTERM`).
- Khi nhận tín hiệu tắt:
  1. Gọi `server.Shutdown(shutdownCtx)` để dừng tiếp nhận request mới và chờ các request đang xử lý hoàn thành trong tối đa 5 giây.
  2. Dừng `WorkerPool.Stop()` để xử lý nốt các sự kiện còn đọng lại trong kênh.
  3. Đóng tiến trình an toàn không thất thoát dữ liệu.

#### Bước 7: Xây Dựng Công Cụ Dòng Lệnh (`cmd/cli/main.go`) & Bộ Kiểm Thử Tự Động
- Công cụ CLI dùng chung các package `internal/model`, `internal/repository`, và `internal/service`. Cho phép người dùng chạy:
  ```bash
  go run cmd/cli/main.go create "Viết tài liệu Go" "Tài liệu chi tiết 6 giai đoạn" HIGH
  go run cmd/cli/main.go list
  ```
- Viết Table-Driven Tests cho toàn bộ các tầng:
  - `internal/model/task_test.go`: Kiểm thử logic validation.
  - `internal/repository/repository_test.go`: Kiểm thử đa luồng và race condition.
  - `internal/handler/task_handler_test.go`: Sử dụng `net/http/httptest` để kiểm tra toàn bộ mã trạng thái HTTP (200, 201, 400, 404) mà không cần mở cổng mạng thật.
  - Chạy kiểm thử nghiêm ngặt:
    ```bash
    go test -v -race ./...
    ```

---

## Tài Nguyên Học Tập Tiếp Theo

Sau khi hoàn thành 6 giai đoạn trên và làm chủ dự án TaskManager, bạn đã có một nền tảng vững chắc để tiếp tục chinh phục các chủ đề nâng cao trong hệ sinh thái Go:

1. **Generics trong Go (Từ Go 1.18+)**:
   - Sử dụng Type Parameters (`[T any, K comparable]`) để viết các cấu trúc dữ liệu tổng quát (ví dụ: Generic Cache, Thread-Safe Map, LinkedList) mà không mất đi tính an toàn kiểu dữ liệu.
2. **Làm Việc Với Cơ Sở Dữ Liệu SQL**:
   - Thư viện chuẩn `database/sql` và cơ chế Connection Pooling (`SetMaxOpenConns`, `SetMaxIdleConns`).
   - Sử dụng các thư viện chuẩn công nghiệp như `pgx` (PostgreSQL driver hiệu năng cao), `sqlx` hoặc `sqlc` (tạo mã Go an toàn trực tiếp từ câu truy vấn SQL).
3. **Microservices & gRPC**:
   - Giao thức RPC hiệu năng cao dựa trên HTTP/2 và Protocol Buffers (`protobuf`).
4. **Tối Ưu Hóa Hiệu Năng & Hồ Sơ Bộ Nhớ (Profiling)**:
   - Sử dụng công cụ `pprof` (`go tool pprof`) để phân tích vết CPU profiling, Heap memory allocation và Block contention profiling.
   - Hiểu sâu về cơ chế Escape Analysis (biến nào thoát ra Heap, biến nào nằm lại trên Stack) để giảm thiểu áp lực lên Garbage Collector.

---

### Bảng Thuật Ngữ Đối Chiếu Anh - Việt (Technical Glossary)

| Thuật ngữ Tiếng Anh | Khái niệm Tiếng Việt & Ngữ cảnh sử dụng |
| :--- | :--- |
| **Goroutine** | Luồng người dùng siêu nhẹ do Go Runtime quản lý (khởi tạo ~2KB) |
| **Channel** | Kênh giao tiếp và truyền thông điệp an toàn đa luồng theo mô hình CSP |
| **Slice Header** | Cấu trúc mô tả lát cắt mảng gồm 3 trường (`ptr`, `len`, `cap`) |
| **Pointer Receiver** | Phương thức nhận con trỏ struct, cho phép sửa đổi dữ liệu gốc |
| **Value Receiver** | Phương thức nhận bản sao của struct, bảo toàn tính bất biến |
| **Struct Embedding** | Kỹ thuật nhúng struct con vào struct cha, thăng hạng các trường |
| **Implicit Interface** | Thỏa mãn interface ngầm định không cần từ khóa `implements` |
| **Zero Value** | Giá trị khởi tạo mặc định của mọi kiểu dữ liệu trong Go |
| **Error Wrapping** | Bọc lỗi ngữ cảnh sử dụng `%w` trong hàm `fmt.Errorf` |
| **Graceful Shutdown** | Quy trình dừng ứng dụng an toàn, giải phóng tài nguyên trước khi thoát |
| **Data Race** | Xung đột bộ nhớ khi nhiều luồng cùng truy cập dữ liệu không có khóa |
| **Deadlock** | Trạng thái bế tắc khi các luồng đợi nhau vô tận |
