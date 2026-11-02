.PHONY: help build build-server build-cli run-server run-cli test test-race test-coverage clean fmt vet

# Variables
BIN_DIR := bin
SERVER_BIN := $(BIN_DIR)/server
CLI_BIN := $(BIN_DIR)/task-cli
COVERAGE_FILE := coverage.out

## help: Display available Makefile targets with descriptions
help:
	@echo "TaskManager Project Makefile"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## [a-zA-Z_-]+:.*?$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = "## |: "}; {printf "  \033[36m%-16s\033[0m %s\n", $$2, $$3}'

## build: Compile both server and CLI binaries into bin/
build: build-server build-cli

## build-server: Compile the HTTP REST API server binary
build-server:
	@mkdir -p $(BIN_DIR)
	@echo "Building server binary -> $(SERVER_BIN)..."
	go build -v -o $(SERVER_BIN) ./cmd/server

## build-cli: Compile the TaskManager CLI binary
build-cli:
	@mkdir -p $(BIN_DIR)
	@echo "Building CLI binary -> $(CLI_BIN)..."
	go build -v -o $(CLI_BIN) ./cmd/cli

## run-server: Run the HTTP REST API server locally
run-server:
	go run ./cmd/server

## run-cli: Run the CLI tool (pass ARGS="<command>" to specify arguments)
run-cli:
	go run ./cmd/cli $(ARGS)

## test: Run all automated unit and integration tests
test:
	go test -v ./...

## test-race: Run all tests with Go race detector enabled
test-race:
	go test -v -race ./...

## test-coverage: Run tests and generate code coverage analysis
test-coverage:
	go test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -func=$(COVERAGE_FILE)
	@echo "Coverage report generated: $(COVERAGE_FILE)"

## clean: Remove compiled binaries, coverage profiles, and temporary test data
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR) $(COVERAGE_FILE) coverage.html

## fmt: Format all Go source files using gofmt
fmt:
	@echo "Formatting Go source files..."
	gofmt -s -w .

## vet: Run go vet static analysis tool
vet:
	@echo "Running go vet..."
	go vet ./...
