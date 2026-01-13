//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/models"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/database/postgres"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func migrateUsers(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		CREATE TABLE users (
			id UUID PRIMARY KEY,
			username TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL
		);
	`)
	require.NoError(t, err)
}

func TestUserRepository_Create(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	migrateUsers(t, db)

	repo, err := postgres.NewUserRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	user, err := models.NewUserFromDB(models.UserFromDBParams{
		ID:           uuid.New().String(),
		Username:     "Test User",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		err = repo.Create(ctx, user)
		require.NoError(t, err)

		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
		require.NoError(t, err)

		require.Equal(t, 1, count)
	})

	t.Run("user exists", func(t *testing.T) {
		err = repo.Create(ctx, user) // this user was created in previous test
		require.ErrorIs(t, err, services.ErrUserRepoExists)

		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
		require.NoError(t, err)

		require.Equal(t, 1, count)
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	migrateUsers(t, db)

	repo, err := postgres.NewUserRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	user, err := models.NewUserFromDB(models.UserFromDBParams{
		ID:           uuid.New().String(),
		Username:     "Test User",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	require.NoError(t, err)

	err = repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		userFromDB, err := repo.FindByID(ctx, user.ID().String())
		require.NoError(t, err)

		require.Equal(t, user.ID().String(), userFromDB.ID().String())
		require.Equal(t, user.Username(), userFromDB.Username())
		require.Equal(t, user.Email(), userFromDB.Email())
		require.Equal(t, user.PasswordHash(), userFromDB.PasswordHash())
	})

	t.Run("user not found", func(t *testing.T) {
		userFromDB, err := repo.FindByID(ctx, uuid.New().String())
		require.ErrorIs(t, err, services.ErrUserRepoNotFound)

		require.Nil(t, userFromDB)
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	migrateUsers(t, db)

	repo, err := postgres.NewUserRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	user, err := models.NewUserFromDB(models.UserFromDBParams{
		ID:           uuid.New().String(),
		Username:     "Test User",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	require.NoError(t, err)

	err = repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		userFromDB, err := repo.FindByEmail(ctx, "test@example.com")
		require.NoError(t, err)

		require.Equal(t, user.ID().String(), userFromDB.ID().String())
		require.Equal(t, user.Username(), userFromDB.Username())
		require.Equal(t, user.Email(), userFromDB.Email())
		require.Equal(t, user.PasswordHash(), userFromDB.PasswordHash())
	})

	t.Run("user not found", func(t *testing.T) {
		userFromDB, err := repo.FindByEmail(ctx, "notexisting@other.me")
		require.ErrorIs(t, err, services.ErrUserRepoNotFound)

		require.Nil(t, userFromDB)
	})
}

func TestUserRepository_Update(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	migrateUsers(t, db)

	repo, err := postgres.NewUserRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	user, err := models.NewUserFromDB(models.UserFromDBParams{
		ID:           uuid.New().String(),
		Username:     "Test User",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	require.NoError(t, err)

	err = repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		userFromDB, err := repo.FindByID(ctx, user.ID().String())
		require.NoError(t, err)

		newUsername := "new user name"
		err = userFromDB.ChangeUsername(newUsername)
		require.NoError(t, err)

		err = repo.Update(ctx, userFromDB)
		require.NoError(t, err)

		updateUserFromDB, err := repo.FindByID(ctx, user.ID().String())
		require.NoError(t, err)

		require.Equal(t, newUsername, updateUserFromDB.Username().String())
	})

	t.Run("user not found", func(t *testing.T) {
		notExistingUser, err := models.NewUserFromDB(models.UserFromDBParams{
			ID:           uuid.New().String(),
			Username:     "Not Existing User",
			Email:        "no@example.com",
			PasswordHash: "password123",
		})
		require.NoError(t, err)

		err = repo.Update(ctx, notExistingUser)
		require.ErrorIs(t, err, services.ErrUserRepoNotFound)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	migrateUsers(t, db)

	repo, err := postgres.NewUserRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	user, err := models.NewUserFromDB(models.UserFromDBParams{
		ID:           uuid.New().String(),
		Username:     "Test User",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	require.NoError(t, err)

	err = repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		err := repo.Delete(ctx, user.ID().String())
		require.NoError(t, err)

		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
		require.NoError(t, err)

		require.Equal(t, 0, count)
	})

	t.Run("user not found", func(t *testing.T) {
		notExistingUser, err := models.NewUserFromDB(models.UserFromDBParams{
			ID:           uuid.New().String(),
			Username:     "Not Existing User",
			Email:        "no@example.com",
			PasswordHash: "password123",
		})
		require.NoError(t, err)

		err = repo.Delete(ctx, notExistingUser.ID().String())
		require.ErrorIs(t, err, services.ErrUserRepoNotFound)
	})
}
