package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// responseWriterInterceptor intercepts WriteHeader and Write calls to capture status code and size.
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

// Logging returns an HTTP middleware that logs incoming requests with structured slog.
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

// Recovery returns an HTTP middleware that intercepts panics, logs stack traces, and returns 500.
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

// CORS returns an HTTP middleware providing Cross-Origin Resource Sharing headers.
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
