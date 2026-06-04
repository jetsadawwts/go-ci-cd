package handler

import (
	"errors"
	"net/http"

	"github.com/jetsadawwts/go-cicd/internal/delivery/http/viewmodel"
	"github.com/jetsadawwts/go-cicd/internal/domain/entity"
	"github.com/jetsadawwts/go-cicd/internal/usecase"
	"github.com/jetsadawwts/go-cicd/pkg/response"
	"github.com/labstack/echo/v4"
)

// TaskHandler handles HTTP requests for task operations.
type TaskHandler struct {
	useCase *usecase.TaskUseCase
}

// NewTaskHandler creates a new TaskHandler.
func NewTaskHandler(uc *usecase.TaskUseCase) *TaskHandler {
	return &TaskHandler{useCase: uc}
}

// Create handles the creation of a new task.
func (h *TaskHandler) Create(c echo.Context) error {
	var req viewmodel.CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request body"))
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}

	task, err := h.useCase.CreateTask(c.Request().Context(), req.Title, req.Description)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to create task"))
	}

	return c.JSON(http.StatusCreated, response.Success(viewmodel.ToTaskResponse(task)))
}

// GetByID handles retrieving a task by its ID.
func (h *TaskHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, response.Error("id is required"))
	}

	task, err := h.useCase.GetTask(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, response.Error("task not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("failed to get task"))
	}

	return c.JSON(http.StatusOK, response.Success(viewmodel.ToTaskResponse(task)))
}

// List handles retrieving all tasks.
func (h *TaskHandler) List(c echo.Context) error {
	tasks, err := h.useCase.ListTasks(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list tasks"))
	}

	return c.JSON(http.StatusOK, response.Success(viewmodel.ToTaskListResponse(tasks)))
}

// Update handles updating an existing task.
func (h *TaskHandler) Update(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, response.Error("id is required"))
	}

	var req viewmodel.UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request body"))
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}

	task, err := h.useCase.UpdateTask(c.Request().Context(), id, req.Title, req.Description, entity.TaskStatus(req.Status))
	if err != nil {
		if errors.Is(err, entity.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, response.Error("task not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("failed to update task"))
	}

	return c.JSON(http.StatusOK, response.Success(viewmodel.ToTaskResponse(task)))
}

// Delete handles deleting a task by its ID.
func (h *TaskHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, response.Error("id is required"))
	}

	if err := h.useCase.DeleteTask(c.Request().Context(), id); err != nil {
		if errors.Is(err, entity.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, response.Error("task not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete task"))
	}

	return c.NoContent(http.StatusNoContent)
}
