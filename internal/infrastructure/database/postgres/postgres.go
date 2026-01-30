package postgres

import (
	"database/sql"
	"fmt"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/config"

	_ "github.com/lib/pq"
)

func MustConnect(cfg config.PostgresConnection) *sql.DB {
	const op = "postgres.Connect"

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Errorf("%s: open db: %w", op, err))
	}

	err = db.Ping()
	if err != nil {
		panic(fmt.Errorf("%s: ping db: %w", op, err))
	}

	return db
}
