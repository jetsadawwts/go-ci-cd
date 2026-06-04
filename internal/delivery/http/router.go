package http

import (
	"log/slog"

	"github.com/jetsadawwts/go-cicd/internal/delivery/http/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewRouter(taskHandler *handler.TaskHandler, healthHandler *handler.HealthHandler) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			slog.Info("request",
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
			)
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Routes
	e.GET("/health", healthHandler.Health)

	v1 := e.Group("/api/v1")
	v1.GET("/tasks", taskHandler.List)
	v1.GET("/tasks/:id", taskHandler.GetByID)
	v1.POST("/tasks", taskHandler.Create)
	v1.PUT("/tasks/:id", taskHandler.Update)
	v1.DELETE("/tasks/:id", taskHandler.Delete)

	return e
}
