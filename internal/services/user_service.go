package services

import (
	"errors"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/models"
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
	// ErrUserRepoExists is returned by repository if the user already exists there
	ErrUserRepoExists = errors.New("user already exists in the repository")
)

// Application-level errors
var (
	// ErrUserExists is returned by UserService if the user already exists
	ErrUserExists = errors.New("user already exists")

	// ErrUserCreateFailed is returned by UserService if an error occured during creation
	ErrUserCreateFailed = errors.New("failed to create user")
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

// Create creates a new user with the given username, email, and password.
// Returns ErrUserExists if a user with the same identifier exists,
// or ErrUserCreateFailed for other creation errors.
func (us *UserService) Create(username, email, password string) error {
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
