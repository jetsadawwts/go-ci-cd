package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jetsadawwts/go-cicd/internal/domain/entity"
	"github.com/jetsadawwts/go-cicd/internal/infrastructure/database"
	"github.com/jetsadawwts/go-cicd/internal/infrastructure/repository"
)

func getTestDSN() string {
	host := os.Getenv("DB_HOST")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "postgres")
	dbName := getEnvOrDefault("DB_NAME", "taskdb_test")
	sslMode := getEnvOrDefault("DB_SSLMODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbName, sslMode,
	)
}

func getEnvOrDefault(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func setupTestDB(t *testing.T) *repository.PostgresTaskRepository {
	t.Helper()

	if os.Getenv("DB_HOST") == "" {
		t.Skip("DB_HOST not set, skipping integration tests")
	}

	ctx := context.Background()
	dsn := getTestDSN()

	// Run migrations
	migrationsPath := "file://../../../migrations"
	if err := database.RunMigrations(dsn, migrationsPath); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create connection pool
	pool, err := database.NewPostgresPool(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Clean up tasks table before each test
	_, err = pool.Exec(ctx, "DELETE FROM tasks")
	if err != nil {
		t.Fatalf("failed to clean tasks table: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return repository.NewPostgresTaskRepository(pool)
}

func createTestTask() *entity.Task {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &entity.Task{
		ID:          uuid.New().String(),
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "todo",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestPostgresTaskRepository_Create(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	task := createTestTask()
	err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify the task was created
	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Title != task.Title {
		t.Errorf("Title = %v, want %v", got.Title, task.Title)
	}
	if got.Description != task.Description {
		t.Errorf("Description = %v, want %v", got.Description, task.Description)
	}
	if got.Status != task.Status {
		t.Errorf("Status = %v, want %v", got.Status, task.Status)
	}
}

func TestPostgresTaskRepository_GetByID(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	task := createTestTask()
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.ID != task.ID {
		t.Errorf("ID = %v, want %v", got.ID, task.ID)
	}
	if got.Title != task.Title {
		t.Errorf("Title = %v, want %v", got.Title, task.Title)
	}
}

func TestPostgresTaskRepository_GetByID_NotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New().String())
	if err == nil {
		t.Fatal("GetByID() expected error, got nil")
	}
	if err != entity.ErrTaskNotFound {
		t.Errorf("GetByID() error = %v, want %v", err, entity.ErrTaskNotFound)
	}
}

func TestPostgresTaskRepository_GetAll(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Create multiple tasks
	task1 := createTestTask()
	task1.Title = "Task 1"
	task2 := createTestTask()
	task2.Title = "Task 2"
	task2.CreatedAt = task1.CreatedAt.Add(1 * time.Second)
	task2.UpdatedAt = task2.CreatedAt

	if err := repo.Create(ctx, task1); err != nil {
		t.Fatalf("Create() task1 error = %v", err)
	}
	if err := repo.Create(ctx, task2); err != nil {
		t.Fatalf("Create() task2 error = %v", err)
	}

	tasks, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("GetAll() returned %d tasks, want 2", len(tasks))
	}

	// Should be ordered by created_at DESC
	if tasks[0].Title != "Task 2" {
		t.Errorf("First task title = %v, want Task 2", tasks[0].Title)
	}
	if tasks[1].Title != "Task 1" {
		t.Errorf("Second task title = %v, want Task 1", tasks[1].Title)
	}
}

func TestPostgresTaskRepository_Update(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	task := createTestTask()
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the task
	task.Title = "Updated Task"
	task.Description = "Updated Description"
	task.Status = "in_progress"
	task.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

	if err := repo.Update(ctx, task); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Title != "Updated Task" {
		t.Errorf("Title = %v, want Updated Task", got.Title)
	}
	if got.Description != "Updated Description" {
		t.Errorf("Description = %v, want Updated Description", got.Description)
	}
	if got.Status != "in_progress" {
		t.Errorf("Status = %v, want in_progress", got.Status)
	}
}

func TestPostgresTaskRepository_Delete(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	task := createTestTask()
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify task is deleted
	_, err := repo.GetByID(ctx, task.ID)
	if err != entity.ErrTaskNotFound {
		t.Errorf("GetByID() after delete error = %v, want %v", err, entity.ErrTaskNotFound)
	}
}

func TestPostgresTaskRepository_Delete_NotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.Delete(ctx, uuid.New().String())
	if err == nil {
		t.Fatal("Delete() expected error, got nil")
	}
	if err != entity.ErrTaskNotFound {
		t.Errorf("Delete() error = %v, want %v", err, entity.ErrTaskNotFound)
	}
}
