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
			return nil, services.ErrUserNotFound
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

//func (ur *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (ur *UserRepository) Update(ctx context.Context, u *models.User) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (ur *UserRepository) Delete(ctx context.Context, id string) error {
//	//TODO implement me
//	panic("implement me")
//}
