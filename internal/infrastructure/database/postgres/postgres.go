package postgres

import (
	"database/sql"
	"fmt"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/config"

	_ "github.com/lib/pq"
)

func DSN(cfg config.PostgresConnection) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)
}

// MustConnect opens a connection to Postgres database with provided configuration.
// Panics if an error occurred
func MustConnect(cfg config.PostgresConnection) *sql.DB {
	const op = "postgres.Connect"

	dsn := DSN(cfg)

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

func Close(db *sql.DB) error {
	if err := db.Close(); err != nil {
		return fmt.Errorf("unable to close the db: %w", err)
	}

	return nil
}
