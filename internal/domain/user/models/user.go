package models

import (
	"errors"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/vo"
	"github.com/google/uuid"
)

// User represents a system user.
// Each user has a unique ID and Email. Password is stored as a hashed value.
type User struct {
	id           uuid.UUID
	username     vo.Username
	email        vo.Email
	passwordHash vo.Password
}

// ID returns the user's unique identifier.
func (u *User) ID() uuid.UUID { return u.id }

// Username returns the user's username value object.
func (u *User) Username() vo.Username { return u.username }

// Email returns the user's email value object.
func (u *User) Email() vo.Email { return u.email }

// PasswordHash returns the user's password hash value object.
func (u *User) PasswordHash() vo.Password { return u.passwordHash }

var (
	// ErrUserIDInvalid indicates that a provided user ID is not a valid UUID.
	ErrUserIDInvalid = errors.New("user ID is invalid")
)

// NewUser creates a new User from raw string inputs.
// It validates the username, email, and password, converts them into value objects,
// hashes the password, and generates a new UUID for the user
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
		id:           uuid.New(),
		username:     usernameVO,
		email:        emailVO,
		passwordHash: passwordVO,
	}, nil
}

// UserFromDBParams contains raw user data loaded from the database.
//
// This struct is used as an input for constructing a domain User
// from persisted storage. All fields represent database values
// and may require validation or transformation before being used
// inside the domain model.
//
// Password must contain a raw password
type UserFromDBParams struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
}

// NewUserFromDB creates a new User with a specified UUID.
// It validates the ID and other user fields, converts them into value objects,
// and hashes the password. Returns ErrUserIDInvalid if the UUID is invalid.
func NewUserFromDB(p UserFromDBParams) (*User, error) {
	usernameVO, err := vo.NewUsername(p.Username)
	if err != nil {
		return nil, err
	}

	emailVO, err := vo.NewEmail(p.Email)
	if err != nil {
		return nil, err
	}

	passwordVO := vo.NewPasswordFromHash([]byte(p.PasswordHash))

	parsedID, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, ErrUserIDInvalid
	}

	user := &User{
		id:           parsedID,
		username:     usernameVO,
		email:        emailVO,
		passwordHash: passwordVO,
	}

	return user, nil
}

// ChangeUsername updates the user's username after validating it.
// Returns an error if the new username is invalid.
func (u *User) ChangeUsername(new string) error {
	newUsernameVO, err := vo.NewUsername(new)
	if err != nil {
		return err
	}

	u.username = newUsernameVO
	return nil
}

// ChangeEmail updates the user's email after validating it.
// Returns an error if the new email is invalid.
func (u *User) ChangeEmail(new string) error {
	newEmailVO, err := vo.NewEmail(new)
	if err != nil {
		return err
	}

	u.email = newEmailVO

	return nil
}

// ChangePassword updates the user's password.
// The old password is verified in this method.
// Returns an error if verification fails or the new password is invalid.
func (u *User) ChangePassword(old, new string) error {
	err := u.passwordHash.Verify(old)
	if err != nil {
		return err
	}

	newPasswordVO, err := vo.NewPassword(new)
	if err != nil {
		return err
	}

	u.passwordHash = newPasswordVO
	return nil
}
