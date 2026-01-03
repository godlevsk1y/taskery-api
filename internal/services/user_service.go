package services

import (
	"errors"
	"fmt"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/models"
	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/vo"
)

// UserService is a service that handles user operations.
type UserService struct {
	usersRepo     UserRepository
	tokenProvider TokenProvider
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

// TokenProvider defines the interface for generating authentication tokens.
type TokenProvider interface {
	Generate(userID string) (string, error)
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

	// ErrUserRegisterFailed is returned by UserService if an internal error occured during creation
	ErrUserRegisterFailed = errors.New("failed to create user")

	// ErrUserLoginFailed is returned by UserService if an internal error occured during login
	ErrUserLoginFailed = errors.New("failed to login")

	// ErrUserChangeEmailFailed is returned by UserService if an internal error occured during email editing
	ErrUserChangeEmailFailed = errors.New("failed to change email")

	// ErrUserChangePasswordFailed is returned by UserService if an internal error occured during password editing
	ErrUserChangePasswordFailed = errors.New("failed to change password")

	// ErrUserDeleteFailed is returned by UserService if an internal error occured during deletion
	ErrUserDeleteFailed = errors.New("failed to delete user")

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
func NewUserService(usersRepo UserRepository, tokenProvider TokenProvider) (*UserService, error) {
	if usersRepo == nil || tokenProvider == nil {
		return nil, ErrUserRepositoryNil
	}

	return &UserService{
		usersRepo:     usersRepo,
		tokenProvider: tokenProvider,
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

		return ErrUserRegisterFailed
	}

	return nil
}

// Login authenticates a user using the given email and password.
// If authentication is successful, it returns a token representing the session.
//
// If no user exists with the given email, Login returns ErrUserNotFound.
// If the password does not match, Login returns ErrUserUnauthorized.
// If token generation or repository access fails, Login returns ErrUserLoginFailed.
func (us *UserService) Login(email, password string) (string, error) {
	user, err := us.usersRepo.FindByEmail(email)
	if errors.Is(err, ErrUserRepoNotFound) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrUserLoginFailed, err)
	}

	err = user.PasswordHash().Verify(password)
	if errors.Is(err, vo.ErrPassowrdNotMatch) {
		return "", ErrUserUnauthorized
	}
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrUserLoginFailed, err)
	}

	token, err := us.tokenProvider.Generate(user.ID().String())
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrUserLoginFailed, err)
	}

	return token, nil
}

// ChangeUsername changes the username of the user with the given id.
//
// The operation verifies the provided password before applying the change.
// If the user does not exist, it returns ErrUserNotFound.
// If the password is invalid, it returns ErrUserUnauthorized.
// If updating the user fails, it returns an error wrapping ErrUserChangeEmailFailed.
//
// On success, ChangeUsername returns nil.
func (us *UserService) ChangeUsername(id, newUsername, password string) error {
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

	if err := user.ChangeUsername(newUsername); err != nil {
		return err
	}

	if err := us.usersRepo.Update(user); err != nil {
		return fmt.Errorf("%w: %s", ErrUserChangeEmailFailed, err)
	}

	return nil
}

// ChangeEmail changes the email address of the user with the given id.
//
// The operation verifies the provided password and ensures that the newEmail
// is not already in use by another user.
// If the user does not exist, it returns ErrUserNotFound.
// If the password is invalid, it returns ErrUserUnauthorized.
// If the new email is already taken, it returns ErrEmailAlreadyTaken.
// If any repository operation fails, it returns an error wrapping
// ErrUserChangeEmailFailed.
//
// On success, ChangeEmail returns nil.
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

// ChangePassword updates the password of the user with the given id.
//
// The operation verifies the old password before applying the new one.
// If the user does not exist, it returns ErrUserNotFound.
// If changing the password fails, it returns the corresponding error from
// the user object.
// If updating the user in the repository fails, it returns an error wrapping
// ErrUserChangePasswordFailed.
//
// On success, ChangePassword returns nil.
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

// Delete deletes the user with the given id after verifying the provided password.
//
// Delete returns ErrUserNotFound if the user does not exist.
// It returns ErrUserUnauthorized if the password does not match.
// If the deletion fails for any other reason, Delete returns ErrUserDeleteFailed.
// On success, Delete returns nil.
func (us *UserService) Delete(id, password string) error {
	user, err := us.usersRepo.FindByID(id)
	if errors.Is(err, ErrUserRepoNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrUserDeleteFailed, err)
	}

	err = user.PasswordHash().Verify(password)
	if errors.Is(err, vo.ErrPassowrdNotMatch) {
		return ErrUserUnauthorized
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrUserDeleteFailed, err)
	}

	if err := us.usersRepo.Delete(id); err != nil {
		return fmt.Errorf("%w: %s", ErrUserDeleteFailed, err)
	}

	return nil
}
