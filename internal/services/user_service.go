package services

import (
	"errors"
	"fmt"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/models"
	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/vo"
)

// UserService is a service that handles user operations.
type UserService struct {
	usersRepo UserRepository
}

// UserRepository defines the methods for managing user data in a persistent storage.
// It provides basic CRUD operations for the User model.
type UserRepository interface {
	// Create saves a new user in the repository.
	// Returns an error if the operation fails.
	Create(u *models.User) error

	// FindByID retrieves a user by its unique identifier.
	// Returns the user and nil error if found, otherwise returns nil and an error.
	FindByID(id string) (*models.User, error)

	// FindByEmail retrieves a user by their email address.
	// Returns the user and nil error if found, otherwise returns nil and an error.
	FindByEmail(email string) (*models.User, error)

	// Update modifies an existing user's data in the repository.
	// Returns an error if the operation fails or the user does not exist.
	Update(u *models.User) error

	// Delete removes a user from the repository by their unique identifier.
	// Returns an error if the operation fails or the user does not exist.
	Delete(id string) error
}

// ErrUserRepositoryNil is an error that indicates that the user repository
// that is passed to NewUserService is nil.
var ErrUserRepositoryNil = errors.New("user repository is nil")

// Repository-level errors
var (
	// ErrUserRepoExists is returned by repository if the user
	// that is to be added to the repository already exists there
	ErrUserRepoExists = errors.New("user already exists in the repository")

	// ErrUserRepoNotFound is returned by repository if the user was not found there
	ErrUserRepoNotFound = errors.New("user was not found in the repository")
)

// Application-level errors
var (
	// ErrUserExists is returned by UserService if the user that is to be added already exists
	ErrUserExists = errors.New("user already exists")

	// ErrUserNotFound is returned by UserService if the user was not found in the repository
	ErrUserNotFound = errors.New("user was not found")

	// ErrUserCreateFailed is returned by UserService if an internal error occured during creation
	ErrUserCreateFailed = errors.New("failed to create user")

	// ErrUserChangeEmailFailed is returned by UserService if an internal error occured during email editing
	ErrUserChangeEmailFailed = errors.New("failed to change email")

	// ErrUserChangePasswordFailed is returned by UserService if an internal error occured during password editing
	ErrUserChangePasswordFailed = errors.New("failed to change password")

	// ErrUserUnauthorized is returned by UserService
	// if the user does not have the necessary permissions to perform the operation.
	ErrUserUnauthorized = errors.New("unauthorized access")

	// ErrEmailAlreadyTaken is returned by UserService
	// if the email that is to change the old one is already taken.
	ErrEmailAlreadyTaken = errors.New("email is already taken")
)

// NewUserService creates a new instance of UserService with
// given user repository. In case the given repository is nil,
// NewUserService returns nil and an error.
func NewUserService(usersRepo UserRepository) (*UserService, error) {
	if usersRepo == nil {
		return nil, ErrUserRepositoryNil
	}

	return &UserService{
		usersRepo: usersRepo,
	}, nil
}

// Register creates a new user with the given username, email, and password.
// Returns ErrUserExists if a user with the same identifier exists,
// or ErrUserCreateFailed for other creation errors.
func (us *UserService) Register(username, email, password string) error {
	user, err := models.NewUser(username, email, password)
	if err != nil {
		return err
	}

	if err := us.usersRepo.Create(user); err != nil {
		if errors.Is(err, ErrUserRepoExists) {
			return ErrUserExists
		}

		return ErrUserCreateFailed
	}

	return nil
}

// ChangeEmail updates the email address of an existing user.
//
// The method performs the following steps:
//   - Retrieves the user by ID.
//   - Verifies the provided password to ensure the user is authorized.
//   - Checks that the new email is not already in use.
//   - Updates the user's email and persists the change.
//
// Returns:
//   - ErrUserNotFound if the user with the given ID does not exist.
//   - ErrUserUnauthorized if the provided password is incorrect.
//   - ErrEmailAlreadyTaken if the new email is already associated with another user.
//   - ErrUserChangeEmailFailed for any internal or repository-related errors.
func (us *UserService) ChangeEmail(id, newEmail, password string) error {
	user, err := us.usersRepo.FindByID(id)
	if errors.Is(err, ErrUserRepoNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrUserChangeEmailFailed, err)
	}

	err = user.PasswordHash().Verify(password)
	if errors.Is(err, vo.ErrPassowrdNotMatch) {
		return ErrUserUnauthorized
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrUserChangeEmailFailed, err)
	}

	_, err = us.usersRepo.FindByEmail(newEmail)
	if err == nil {
		return ErrEmailAlreadyTaken
	}
	if !errors.Is(err, ErrUserNotFound) {
		return fmt.Errorf("%w: %s", ErrUserChangeEmailFailed, err)
	}

	if err := user.ChangeEmail(newEmail); err != nil {
		return err
	}

	if err := us.usersRepo.Update(user); err != nil {
		return fmt.Errorf("%w: %s", ErrUserChangeEmailFailed, err)
	}

	return nil
}

// ChangePassword updates the user's password.
//
// The method retrieves a user by the given ID, verifies the current (old)
// password, and replaces it with a new one. If the user does not exist,
// ErrUserNotFound is returned. Any repository-related failure during
// retrieval or update is wrapped with ErrUserChangePasswordFailed.
//
// Parameters:
//   - id:   The unique identifier of the user.
//   - old:  The user's current password (used for verification).
//   - new:  The new password to be set.
//
// Returns:
//   - ErrUserNotFound if no user with the given ID exists.
//   - ErrUserChangePasswordFailed if a repository operation fails.
//   - Any error returned by user.ChangePassword (e.g. invalid current
//     password or new password validation failure).
//   - nil on success.
func (us *UserService) ChangePassword(id, old, new string) error {
	user, err := us.usersRepo.FindByID(id)
	if errors.Is(err, ErrUserRepoNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrUserChangePasswordFailed, err)
	}

	if err := user.ChangePassword(old, new); err != nil {
		return err
	}

	if err := us.usersRepo.Update(user); err != nil {
		return fmt.Errorf("%w: %s", ErrUserChangePasswordFailed, err)
	}

	return nil
}
