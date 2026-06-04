package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	db *pgxpool.Pool
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health checks the health of the service and its dependencies.
func (h *HealthHandler) Health(c echo.Context) error {
	status := "ok"

	if err := h.db.Ping(c.Request().Context()); err != nil {
		status = "error"
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status":  status,
			"service": "task-service",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  status,
		"service": "task-service",
	})
}
