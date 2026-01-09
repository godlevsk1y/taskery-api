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

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

//func (ur *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
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
