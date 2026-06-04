package repository

import (
	"context"

	"github.com/jetsadawwts/go-cicd/internal/domain/entity"
)

// TaskRepository defines the interface for task persistence operations.
type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) error
	GetByID(ctx context.Context, id string) (*entity.Task, error)
	GetAll(ctx context.Context) ([]*entity.Task, error)
	Update(ctx context.Context, task *entity.Task) error
	Delete(ctx context.Context, id string) error
}
