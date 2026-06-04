package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jetsadawwts/go-cicd/internal/domain/entity"
	"github.com/jetsadawwts/go-cicd/internal/domain/repository"
)

// TaskUseCase handles business logic for task operations.
type TaskUseCase struct {
	repo repository.TaskRepository
}

// NewTaskUseCase creates a new TaskUseCase with the given repository.
func NewTaskUseCase(repo repository.TaskRepository) *TaskUseCase {
	return &TaskUseCase{repo: repo}
}

// CreateTask creates a new task, validates it, and persists it via the repository.
func (uc *TaskUseCase) CreateTask(ctx context.Context, title, description string) (*entity.Task, error) {
	task := entity.NewTask(title, description)
	if err := task.Validate(); err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

// GetTask retrieves a task by its ID.
func (uc *TaskUseCase) GetTask(ctx context.Context, id string) (*entity.Task, error) {
	task, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// ListTasks retrieves all tasks.
func (uc *TaskUseCase) ListTasks(ctx context.Context) ([]*entity.Task, error) {
	tasks, err := uc.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, nil
}

// UpdateTask updates an existing task's fields, validates, and persists the changes.
func (uc *TaskUseCase) UpdateTask(ctx context.Context, id, title, description string, status entity.TaskStatus) (*entity.Task, error) {
	task, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task for update: %w", err)
	}

	if !entity.IsValidStatus(status) {
		return nil, fmt.Errorf("%w: invalid status %q", entity.ErrInvalidTask, status)
	}

	task.Title = title
	task.Description = description
	task.Status = status
	task.UpdatedAt = time.Now()

	if err := task.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	return task, nil
}

// DeleteTask removes a task by its ID.
func (uc *TaskUseCase) DeleteTask(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}
