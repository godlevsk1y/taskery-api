package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupPostgres returns a db and a cleanup func
func setupPostgres(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	ctx := context.Background()
	dbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "test_db",
			},
			HostConfigModifier: func(config *container.HostConfig) {
				config.AutoRemove = true
			},
			WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(15 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)

	host, err := dbContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := dbContainer.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	dsn := fmt.Sprintf(
		"postgres://test:test@%s:%s/test_db?sslmode=disable",
		host,
		mappedPort.Port(),
	)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	cleanup := func() {
		_ = db.Close()
		_ = dbContainer.Terminate(ctx)
	}

	return db, cleanup
}
