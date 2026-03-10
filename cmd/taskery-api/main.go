package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/auth/jwt"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/config"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/database"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/database/postgres"
	v1 "github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/go-playground/validator/v10"
)

// @title Taskery API
// @version 1.0
// @description API documentation for Taskery
// @BasePath /v1

func main() {
	cfg := config.MustLoad()

	logger := setupLogger(cfg.Environment)

	logger.Info("Starting taskery-api...")

	db := postgres.MustConnect(cfg.PostgresConnection)

	err := database.RunMigrations(postgres.DSN(cfg.PostgresConnection))
	if err != nil {
		logger.Error("Failed to apply migrations", slog.Any("err", err))
		os.Exit(-1)
	}

	logger.Info("Connection to database succeeded.")

	userRepo, err := postgres.NewUserRepository(db)
	if err != nil {
		logger.Error("Failed to init user repository", slog.Any("err", err))
		os.Exit(-1)
	}

	taskRepo, err := postgres.NewTaskRepository(db)
	if err != nil {
		logger.Error("Failed to init task repository", slog.Any("err", err))
		os.Exit(-1)
	}

	logger.Info("Repositories initialization succeeded.")

	jwtProvider := jwt.NewProvider([]byte(cfg.JWT.Secret), cfg.JWT.TTL, cfg.JWT.Issuer)

	userSvc, err := services.NewUserService(userRepo, jwtProvider)
	if err != nil {
		logger.Error("Failed to init user service", slog.Any("err", err))
		os.Exit(-1)
	}

	taskSvc, err := services.NewTaskService(taskRepo)
	if err != nil {
		logger.Error("Failed to init task service", slog.Any("err", err))
		os.Exit(-1)
	}

	vld := validator.New()

	router := v1.NewRouter(v1.RouterOptions{
		UserService:   userSvc,
		TaskService:   taskSvc,
		Logger:        logger,
		TokenProvider: jwtProvider,
		Validator:     vld,
		Timeout:       cfg.HTTPServer.Timeout,
	})

	logger.Info(
		"Launching the server...",
		slog.String("address", cfg.HTTPServer.Address),
	)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("Server not running", slog.Any("err", err))
		}
	}()

	logger.Info("Server started")

	<-done

	logger.Info("Launching server shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown the server")

		return
	}

	if err := db.Close(); err != nil {
		logger.Error(
			"Unable to close the connection to Postgres database",
			slog.Any("err", err),
		)
	} else {
		logger.Info("Successfully closed Postgres database connection")
	}

	logger.Info("Server stopped")
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case "local":
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

	case "dev":
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "production":
	default:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return logger
}
