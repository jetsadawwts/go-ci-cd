package usecase_test

import (
	"context"
	"sync"
	"testing"

	"github.com/jetsadawwts/go-cicd/internal/domain/entity"
	"github.com/jetsadawwts/go-cicd/internal/usecase"
)

// mockTaskRepository is an in-memory implementation of repository.TaskRepository for testing.
type mockTaskRepository struct {
	tasks map[string]*entity.Task
	mu    sync.RWMutex
}

func newMockTaskRepository() *mockTaskRepository {
	return &mockTaskRepository{
		tasks: make(map[string]*entity.Task),
	}
}

func (m *mockTaskRepository) Create(_ context.Context, task *entity.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) GetByID(_ context.Context, id string) (*entity.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	task, ok := m.tasks[id]
	if !ok {
		return nil, entity.ErrTaskNotFound
	}
	return task, nil
}

func (m *mockTaskRepository) GetAll(_ context.Context) ([]*entity.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tasks := make([]*entity.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m *mockTaskRepository) Update(_ context.Context, task *entity.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tasks[task.ID]; !ok {
		return entity.ErrTaskNotFound
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tasks[id]; !ok {
		return entity.ErrTaskNotFound
	}
	delete(m.tasks, id)
	return nil
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		wantErr     bool
	}{
		{
			name:        "success",
			title:       "Test Task",
			description: "A test task description",
			wantErr:     false,
		},
		{
			name:        "empty title error",
			title:       "",
			description: "Description without title",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockTaskRepository()
			uc := usecase.NewTaskUseCase(repo)
			ctx := context.Background()

			task, err := uc.CreateTask(ctx, tt.title, tt.description)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if task.Title != tt.title {
				t.Errorf("expected title %q, got %q", tt.title, task.Title)
			}
			if task.Description != tt.description {
				t.Errorf("expected description %q, got %q", tt.description, task.Description)
			}
			if task.Status != entity.StatusTodo {
				t.Errorf("expected status %q, got %q", entity.StatusTodo, task.Status)
			}
			if task.ID == "" {
				t.Error("expected non-empty ID")
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	tests := []struct {
		setup   func(uc *usecase.TaskUseCase) string // returns task ID
		name    string
		wantErr bool
	}{
		{
			name: "success",
			setup: func(uc *usecase.TaskUseCase) string {
				task, _ := uc.CreateTask(context.Background(), "Test", "Desc")
				return task.ID
			},
			wantErr: false,
		},
		{
			name: "not found",
			setup: func(_ *usecase.TaskUseCase) string {
				return "non-existent-id"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockTaskRepository()
			uc := usecase.NewTaskUseCase(repo)
			ctx := context.Background()

			id := tt.setup(uc)
			task, err := uc.GetTask(ctx, id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if task.ID != id {
				t.Errorf("expected ID %q, got %q", id, task.ID)
			}
		})
	}
}

func TestListTasks(t *testing.T) {
	tests := []struct {
		name      string
		seedCount int
		wantCount int
	}{
		{
			name:      "empty list",
			seedCount: 0,
			wantCount: 0,
		},
		{
			name:      "with items",
			seedCount: 3,
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockTaskRepository()
			uc := usecase.NewTaskUseCase(repo)
			ctx := context.Background()

			for i := 0; i < tt.seedCount; i++ {
				_, err := uc.CreateTask(ctx, "Task", "Description")
				if err != nil {
					t.Fatalf("failed to seed task: %v", err)
				}
			}

			tasks, err := uc.ListTasks(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tasks) != tt.wantCount {
				t.Errorf("expected %d tasks, got %d", tt.wantCount, len(tasks))
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(uc *usecase.TaskUseCase) string
		title   string
		desc    string
		status  entity.TaskStatus
		wantErr bool
	}{
		{
			name: "success",
			setup: func(uc *usecase.TaskUseCase) string {
				task, _ := uc.CreateTask(context.Background(), "Original", "Desc")
				return task.ID
			},
			title:   "Updated Title",
			desc:    "Updated Description",
			status:  entity.StatusInProgress,
			wantErr: false,
		},
		{
			name: "not found",
			setup: func(_ *usecase.TaskUseCase) string {
				return "non-existent-id"
			},
			title:   "Title",
			desc:    "Desc",
			status:  entity.StatusTodo,
			wantErr: true,
		},
		{
			name: "invalid status",
			setup: func(uc *usecase.TaskUseCase) string {
				task, _ := uc.CreateTask(context.Background(), "Task", "Desc")
				return task.ID
			},
			title:   "Title",
			desc:    "Desc",
			status:  entity.TaskStatus("invalid"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockTaskRepository()
			uc := usecase.NewTaskUseCase(repo)
			ctx := context.Background()

			id := tt.setup(uc)
			task, err := uc.UpdateTask(ctx, id, tt.title, tt.desc, tt.status)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if task.Title != tt.title {
				t.Errorf("expected title %q, got %q", tt.title, task.Title)
			}
			if task.Description != tt.desc {
				t.Errorf("expected description %q, got %q", tt.desc, task.Description)
			}
			if task.Status != tt.status {
				t.Errorf("expected status %q, got %q", tt.status, task.Status)
			}
		})
	}
}

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		setup   func(uc *usecase.TaskUseCase) string
		name    string
		wantErr bool
	}{
		{
			name: "success",
			setup: func(uc *usecase.TaskUseCase) string {
				task, _ := uc.CreateTask(context.Background(), "To Delete", "Desc")
				return task.ID
			},
			wantErr: false,
		},
		{
			name: "not found",
			setup: func(_ *usecase.TaskUseCase) string {
				return "non-existent-id"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockTaskRepository()
			uc := usecase.NewTaskUseCase(repo)
			ctx := context.Background()

			id := tt.setup(uc)
			err := uc.DeleteTask(ctx, id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify task is actually deleted
			_, err = uc.GetTask(ctx, id)
			if err == nil {
				t.Error("expected error when getting deleted task, got nil")
			}
		})
	}
}
