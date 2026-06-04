package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	deliveryhttp "github.com/jetsadawwts/go-cicd/internal/delivery/http"
	"github.com/jetsadawwts/go-cicd/internal/delivery/http/handler"
	"github.com/jetsadawwts/go-cicd/internal/infrastructure/config"
	"github.com/jetsadawwts/go-cicd/internal/infrastructure/database"
	"github.com/jetsadawwts/go-cicd/internal/infrastructure/repository"
	"github.com/jetsadawwts/go-cicd/internal/usecase"
)

var version = "dev"

func main() {
	slog.Info("starting application", "version", version)

	cfg := config.Load()

	ctx := context.Background()

	if err := database.RunMigrations(cfg.Database.DSN(), "file://migrations"); err != nil {
		slog.Warn("failed to run database migrations", "error", err)
	}

	pool, err := database.NewPostgresPool(ctx, cfg.Database.DSN())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewPostgresTaskRepository(pool)
	taskUseCase := usecase.NewTaskUseCase(repo)
	taskHandler := handler.NewTaskHandler(taskUseCase)
	healthHandler := handler.NewHealthHandler(pool)

	router := deliveryhttp.NewRouter(taskHandler, healthHandler)

	// Start server
	go func() {
		slog.Info("server starting", "port", cfg.Server.Port)
		if err := router.Start(":" + cfg.Server.Port); err != nil && err != http.ErrServerClosed {
			slog.Error("shutting down the server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("received shutdown signal")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := router.Shutdown(ctxShutdown); err != nil {
		slog.Error("fatal server shutdown", "error", err)
	}

	slog.Info("server stopped gracefully")
}
