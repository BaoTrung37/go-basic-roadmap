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
	// Initialize structured logger
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

	// 1. Initialize Storage Repository
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

	// 2. Initialize and Start Concurrency Worker Pool
	workerPool := worker.NewWorkerPool(cfg.WorkerPoolSize, cfg.WorkerQueueSize, logger)
	poolCtx, cancelPool := context.WithCancel(context.Background())
	defer cancelPool()
	workerPool.Start(poolCtx)

	// 3. Initialize Service Layer
	taskService := service.NewTaskService(repo, workerPool, logger)

	// 4. Initialize HTTP Handler & Router
	mux := http.NewServeMux()
	taskHandler := handler.NewTaskHandler(taskService, logger)
	taskHandler.RegisterRoutes(mux)

	// 5. Apply Middlewares: CORS -> Recovery -> Logging
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

	// 6. Start Server in Background Goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		logger.Info(fmt.Sprintf("HTTP server listening on %s", cfg.ServerPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	// 7. Await Shutdown Signal
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

	// 8. Graceful Shutdown with Timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server graceful shutdown failed, forcing close", "error", err)
		_ = server.Close()
	} else {
		logger.Info("HTTP server stopped gracefully")
	}

	// Stop background workers and drain remaining events
	workerPool.Stop()
	logger.Info("TaskManager server exited cleanly")
}
