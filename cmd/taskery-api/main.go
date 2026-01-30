package main

import (
	"log/slog"
	"os"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/config"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/database/postgres"
)

func main() {
	cfg := config.MustLoad()

	logger := setupLogger(cfg.Environment)

	logger.Info("starting taskery-api...")

	db := postgres.MustConnect(cfg.PostgresConnection)

	_ = db
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
