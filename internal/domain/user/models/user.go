package models

import (
	"errors"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/vo"
	"github.com/google/uuid"
)

// User is a model that represents a user.
// ID and Email are unique for each user.
type User struct {
	ID           uuid.UUID
	Username     vo.Username
	Email        vo.Email
	PasswordHash vo.Password
}

var (
	ErrUserIDInvalid = errors.New("user ID is invalid")
)

// NewUser creates a new User from raw string inputs.
// It validates and converts the inputs into value objects and hashes the password.
func NewUser(username, email, password string) (*User, error) {
	usernameVO, err := vo.NewUsername(username)
	if err != nil {
		return nil, err
	}

	emailVO, err := vo.NewEmail(email)
	if err != nil {
		return nil, err
	}

	passwordVO, err := vo.NewPassword(password)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           uuid.New(),
		Username:     usernameVO,
		Email:        emailVO,
		PasswordHash: passwordVO,
	}, nil
}

// NewUserWithID creates a new User from raw string inputs and a UUID ID.
// It validates and converts the inputs into value objects and hashes the password.
func NewUserWithID(id, username, email, password string) (*User, error) {
	user, err := NewUser(username, email, password)
	if err != nil {
		return nil, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrUserIDInvalid
	}

	user.ID = parsedID
	return user, nil
}
