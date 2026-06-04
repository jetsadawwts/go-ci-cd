package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/jetsadawwts/go-cicd/internal/delivery/http/handler"
	"github.com/jetsadawwts/go-cicd/internal/domain/entity"
	"github.com/jetsadawwts/go-cicd/internal/usecase"
	"github.com/jetsadawwts/go-cicd/pkg/response"
	"github.com/labstack/echo/v4"
)

// mockTaskRepository implements repository.TaskRepository using an in-memory map.
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
	for _, t := range m.tasks {
		tasks = append(tasks, t)
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

func setupHandler() *handler.TaskHandler {
	repo := newMockTaskRepository()
	uc := usecase.NewTaskUseCase(repo)
	return handler.NewTaskHandler(uc)
}

func TestCreateTask(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		h := setupHandler()

		body := `{"title":"Test Task","description":"A test task"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Create(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		var resp response.Response
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("expected resp.Data to be map[string]interface{}, got %T", resp.Data)
		}
		if data["title"] != "Test Task" {
			t.Errorf("expected title 'Test Task', got '%s'", data["title"])
		}
		if data["status"] != "todo" {
			t.Errorf("expected status 'todo', got '%s'", data["status"])
		}
		if data["id"] == nil || data["id"] == "" {
			t.Error("expected id to be set")
		}
	})

	t.Run("bad request - empty title", func(t *testing.T) {
		h := setupHandler()

		body := `{"title":"","description":"A test task"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Create(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("bad request - invalid json", func(t *testing.T) {
		h := setupHandler()

		body := `{invalid json}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Create(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})
}

func TestGetByID(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		h := setupHandler()

		// First create a task.
		body := `{"title":"Test Task","description":"A test task"}`
		createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()
		createCtx := e.NewContext(createReq, createRec)
		if err := h.Create(createCtx); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		var created response.Response
		if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		data, ok := created.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("expected created.Data to be map[string]interface{}, got %T", created.Data)
		}
		id, ok := data["id"].(string)
		if !ok {
			t.Fatalf("expected data[id] to be string, got %T", data["id"])
		}

		// Now get the task by ID.
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+id, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(id)

		err := h.GetByID(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp response.Response
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		respData, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("expected resp.Data to be map[string]interface{}, got %T", resp.Data)
		}

		if respData["id"] != id {
			t.Errorf("expected id '%s', got '%s'", id, respData["id"])
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := setupHandler()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/nonexistent", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent")

		err := h.GetByID(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}

func TestList(t *testing.T) {
	e := echo.New()
	h := setupHandler()

	// Create two tasks.
	for _, title := range []string{"Task 1", "Task 2"} {
		body := `{"title":"` + title + `","description":"desc"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := h.Create(c); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	// List all tasks.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.List(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp response.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected resp.Data to be map[string]interface{}, got %T", resp.Data)
	}

	totalVal, ok := data["total"].(float64)
	if !ok {
		t.Fatalf("expected data[total] to be float64, got %T", data["total"])
	}
	total := int(totalVal)
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}

	tasks, ok := data["tasks"].([]interface{})
	if !ok {
		t.Fatalf("expected data[tasks] to be []interface{}, got %T", data["tasks"])
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestUpdate(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		h := setupHandler()

		// Create a task.
		body := `{"title":"Original","description":"original desc"}`
		createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()
		createCtx := e.NewContext(createReq, createRec)
		if err := h.Create(createCtx); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		var created response.Response
		if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		data, ok := created.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("expected created.Data to be map[string]interface{}, got %T", created.Data)
		}
		id, ok := data["id"].(string)
		if !ok {
			t.Fatalf("expected data[id] to be string, got %T", data["id"])
		}

		// Update the task.
		updateBody := `{"title":"Updated","description":"updated desc","status":"done"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/"+id, bytes.NewBufferString(updateBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(id)

		err := h.Update(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp response.Response
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		respData, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("expected resp.Data to be map[string]interface{}, got %T", resp.Data)
		}

		if respData["title"] != "Updated" {
			t.Errorf("expected title 'Updated', got '%s'", respData["title"])
		}
		if respData["status"] != "done" {
			t.Errorf("expected status 'done', got '%s'", respData["status"])
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := setupHandler()

		updateBody := `{"title":"Updated","description":"updated desc"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/nonexistent", bytes.NewBufferString(updateBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent")

		err := h.Update(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}

func TestDelete(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		h := setupHandler()

		// Create a task.
		body := `{"title":"To Delete","description":"will be deleted"}`
		createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()
		createCtx := e.NewContext(createReq, createRec)
		if err := h.Create(createCtx); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		var created response.Response
		if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		data, ok := created.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("expected created.Data to be map[string]interface{}, got %T", created.Data)
		}
		id, ok := data["id"].(string)
		if !ok {
			t.Fatalf("expected data[id] to be string, got %T", data["id"])
		}

		// Delete the task.
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+id, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(id)

		err := h.Delete(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}

		// Verify it's deleted.
		getReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+id, nil)
		getRec := httptest.NewRecorder()
		getCtx := e.NewContext(getReq, getRec)
		getCtx.SetParamNames("id")
		getCtx.SetParamValues(id)

		err = h.GetByID(getCtx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if getRec.Code != http.StatusNotFound {
			t.Errorf("expected status %d after delete, got %d", http.StatusNotFound, getRec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := setupHandler()

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/nonexistent", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent")

		err := h.Delete(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}
