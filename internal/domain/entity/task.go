package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

// Custom errors for task operations.
var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidTask  = errors.New("invalid task")
)

// Task represents a task entity in the domain.
type Task struct {
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
}

// NewTask creates a new task with a generated UUID, status set to Todo, and timestamps set to now.
func NewTask(title, description string) *Task {
	now := time.Now()
	return &Task{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Status:      StatusTodo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate checks that the task has a non-empty title.
func (t *Task) Validate() error {
	if t.Title == "" {
		return ErrInvalidTask
	}
	return nil
}

// MarkInProgress sets the task status to in-progress and updates the timestamp.
func (t *Task) MarkInProgress() {
	t.Status = StatusInProgress
	t.UpdatedAt = time.Now()
}

// MarkDone sets the task status to done and updates the timestamp.
func (t *Task) MarkDone() {
	t.Status = StatusDone
	t.UpdatedAt = time.Now()
}

// IsValidStatus returns true if the given status is a recognized TaskStatus.
func IsValidStatus(s TaskStatus) bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
