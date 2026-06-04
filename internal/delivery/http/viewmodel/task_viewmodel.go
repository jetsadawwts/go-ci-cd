package viewmodel

import (
	"errors"
	"strings"
	"time"

	"github.com/jetsadawwts/go-cicd/internal/domain/entity"
)

// CreateTaskRequest represents the request body for creating a task.
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateTaskRequest represents the request body for updating a task.
type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// TaskResponse represents the response body for a single task.
type TaskResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TaskListResponse represents the response body for a list of tasks.
type TaskListResponse struct {
	Tasks []TaskResponse `json:"tasks"`
	Total int            `json:"total"`
}

// validStatuses defines allowed task statuses.
var validStatuses = map[string]bool{
	"todo":        true,
	"in_progress": true,
	"done":        true,
}

// ToTaskResponse maps an entity.Task to a TaskResponse.
func ToTaskResponse(t *entity.Task) TaskResponse {
	return TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
}

// ToTaskListResponse maps a slice of entity.Task to a TaskListResponse.
func ToTaskListResponse(tasks []*entity.Task) TaskListResponse {
	responses := make([]TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		responses = append(responses, ToTaskResponse(t))
	}
	return TaskListResponse{
		Tasks: responses,
		Total: len(responses),
	}
}

// Validate validates the CreateTaskRequest.
func (r *CreateTaskRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	return nil
}

// Validate validates the UpdateTaskRequest.
func (r *UpdateTaskRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if r.Status != "" && !validStatuses[r.Status] {
		return errors.New("invalid status: must be one of todo, in_progress, done")
	}
	return nil
}
