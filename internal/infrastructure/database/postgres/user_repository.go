package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/models"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/lib/pq"
)

// UserRepository represents a repository of Taskery users in PostgreSQL database
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository using the provided sql.DB.
// It returns an error if the db argument is nil.
func NewUserRepository(db *sql.DB) (*UserRepository, error) {
	const op = "postgres.UserRepository.NewUserRepository"

	if db == nil {
		return nil, fmt.Errorf("%s: db is nil", op)
	}

	return &UserRepository{db}, nil
}

// Create inserts a new user into the database.
// It stores the user's ID, username, email, and password hash.
//
// If a user with the same unique fields already exists, Create returns
// ErrUserRepoNotFound. Other database errors are returned as-is.
//
// Create does not modify the User object passed in.
func (ur *UserRepository) Create(ctx context.Context, u *models.User) error {
	const op = "postgres.UserRepository.Create"

	const query = `INSERT INTO users(id, username, email, password_hash) VALUES ($1, $2, $3, $4)`

	_, err := ur.db.ExecContext(
		ctx, query,
		u.ID().String(),
		u.Username().String(),
		u.Email().String(),
		u.PasswordHash().String(),
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" { // unique constraint
				return services.ErrUserRepoExists
			}
		}

		return fmt.Errorf("%s: create user: %w", op, err)
	}

	return nil
}

// FindByID looks up a user by its id in the repository.
//
// It returns the corresponding *models.User if the user exists.
// If no user with the given id is found, FindByID returns
// services.ErrUserNotFound.
//
// If a database error occurs while querying or scanning the result,
// or if the user cannot be restored from the persisted data,
// FindByID returns a non-nil error wrapping the underlying failure.
//
// The operation respects the provided context ctx.
func (ur *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	const op = "postgres.UserRepository.FindByID"

	const query = `SELECT id, username, email, password_hash FROM users WHERE id = $1`

	row := ur.db.QueryRowContext(ctx, query, id)

	var (
		userID       string
		email        string
		username     string
		passwordHash string
	)

	err := row.Scan(&userID, &username, &email, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrUserRepoNotFound
		}

		return nil, fmt.Errorf("%s: find by id: %w", op, err)
	}

	user, err := models.NewUserFromDB(models.UserFromDBParams{
		ID:           userID,
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: restore user: %w", op, err)
	}

	return user, nil
}

// FindByEmail looks up a user by its email in the repository.
//
// It returns the corresponding *models.User if the user exists.
// If no user with the given email is found, FindByEmail returns
// services.ErrUserNotFound.
//
// If a database error occurs while querying or scanning the result,
// or if the user cannot be restored from the persisted data,
// FindByID returns a non-nil error wrapping the underlying failure.
//
// The operation respects the provided context ctx.
func (ur *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "postgres.UserRepository.FindByEmail"

	const query = `SELECT id, username, email, password_hash FROM users WHERE email = $1`

	row := ur.db.QueryRowContext(ctx, query, email)

	var (
		userID       string
		userEmail    string
		username     string
		passwordHash string
	)

	err := row.Scan(&userID, &username, &userEmail, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrUserRepoNotFound
		}

		return nil, fmt.Errorf("%s: find by email: %w", op, err)
	}

	user, err := models.NewUserFromDB(models.UserFromDBParams{
		ID:           userID,
		Email:        userEmail,
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: restore user: %w", op, err)
	}

	return user, nil
}

// Update updates the persisted data of the given user u.
//
// It stores the user's current username, email, and password hash
// identified by u.ID.
//
// If no user with the given ID exists, Update returns
// services.ErrUserRepoNotFound.
//
// If a database error occurs while executing the update or determining
// the number of affected rows, Update returns a non-nil error wrapping
// the underlying failure.
//
// The operation respects the provided context ctx.
func (ur *UserRepository) Update(ctx context.Context, u *models.User) error {
	const op = "postgres.UserRepository.Update"

	const query = `UPDATE users SET username = $1, email = $2, password_hash = $3 WHERE id = $4`

	res, err := ur.db.ExecContext(
		ctx,
		query,
		u.Username().String(),
		u.Email().String(),
		u.PasswordHash().String(),
		u.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("%s: update user: %w", op, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: get rows affected: %w", op, err)
	}

	if affected == 0 {
		return services.ErrUserRepoNotFound
	}

	return nil
}

// Delete removes the user with the given id from the repository.
//
// If no user with the given id exists, Delete returns services.ErrUserRepoNotFound.
// Any database or execution error encountered during the operation is returned
// as a non-nil error.
//
// Delete respects the provided context ctx.
func (ur *UserRepository) Delete(ctx context.Context, id string) error {
	const op = "postgres.UserRepository.Delete"

	const query = `DELETE FROM users WHERE id = $1`

	res, err := ur.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: delete user: %w", op, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: get rows affected: %w", op, err)
	}

	if affected == 0 {
		return services.ErrUserRepoNotFound
	}

	return nil
}

var _ services.UserRepository = (*UserRepository)(nil)
