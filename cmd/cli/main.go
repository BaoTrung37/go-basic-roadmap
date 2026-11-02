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

	// Global flags
	var (
		storageType string
		dataPath    string
	)

	defaultDataPath := cfg.DataFilePath
	if defaultDataPath == "" {
		defaultDataPath = "data/tasks.json"
	}

	flag.StringVar(&storageType, "storage", "file", "Storage backend: \"file\" or \"memory\"")
	flag.StringVar(&dataPath, "data", defaultDataPath, "Path to JSON tasks data file (alias: -data-file)")
	flag.StringVar(&dataPath, "data-file", defaultDataPath, "Path to JSON tasks data file (alias: -data)")

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

	// Initialize Repository based on storage configuration
	var repo repository.TaskRepository
	switch strings.ToLower(storageType) {
	case "memory":
		repo = repository.NewMemoryRepository()
	case "file":
		var err error
		repo, err = repository.NewFileRepository(dataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing repository: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid storage type %q (supported: \"file\", \"memory\")\n", storageType)
		os.Exit(1)
	}

	// In CLI mode, worker pool can be nil or background; here we pass nil as CLI is synchronous
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

	// Retrieve existing to keep un-updated fields
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
