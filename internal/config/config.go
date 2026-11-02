package config

import (
	"os"
	"strconv"
)

// Config holds runtime configuration options for the TaskManager application.
type Config struct {
	ServerPort      string
	StorageType     string // "memory" or "file"
	DataFilePath    string
	WorkerPoolSize  int
	WorkerQueueSize int
	LogLevel        string
}

// Load loads configuration from environment variables with sensible defaults.
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
