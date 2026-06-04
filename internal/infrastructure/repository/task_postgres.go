package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jetsadawwts/go-cicd/internal/domain/entity"
)

// PostgresTaskRepository implements repository.TaskRepository using PostgreSQL.
type PostgresTaskRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTaskRepository creates a new PostgresTaskRepository.
func NewPostgresTaskRepository(pool *pgxpool.Pool) *PostgresTaskRepository {
	return &PostgresTaskRepository{pool: pool}
}

// Create inserts a new task into the database.
func (r *PostgresTaskRepository) Create(ctx context.Context, task *entity.Task) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO tasks (id, title, description, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		task.ID, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a task by its ID.
func (r *PostgresTaskRepository) GetByID(ctx context.Context, id string) (*entity.Task, error) {
	var task entity.Task
	err := r.pool.QueryRow(ctx,
		"SELECT id, title, description, status, created_at, updated_at FROM tasks WHERE id = $1",
		id,
	).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

// GetAll retrieves all tasks ordered by creation date descending.
func (r *PostgresTaskRepository) GetAll(ctx context.Context) ([]*entity.Task, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT id, title, description, status, created_at, updated_at FROM tasks ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*entity.Task
	for rows.Next() {
		var task entity.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// Update modifies an existing task in the database.
func (r *PostgresTaskRepository) Update(ctx context.Context, task *entity.Task) error {
	result, err := r.pool.Exec(ctx,
		"UPDATE tasks SET title = $1, description = $2, status = $3, updated_at = $4 WHERE id = $5",
		task.Title, task.Description, task.Status, task.UpdatedAt, task.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return entity.ErrTaskNotFound
	}

	return nil
}

// Delete removes a task from the database by its ID.
func (r *PostgresTaskRepository) Delete(ctx context.Context, id string) error {
	result, err := r.pool.Exec(ctx,
		"DELETE FROM tasks WHERE id = $1",
		id,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return entity.ErrTaskNotFound
	}

	return nil
}
