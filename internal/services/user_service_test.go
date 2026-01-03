package services_test

import (
	"errors"
	"testing"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/models"
	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/vo"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/cyberbrain-dev/taskery-api/internal/services/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewUserService(t *testing.T) {
	tests := []struct {
		name          string
		usersRepo     services.UserRepository
		tokenProvider services.TokenProvider
		wantErr       error
	}{
		{
			name:          "valid repository",
			usersRepo:     new(mocks.UserRepository),
			tokenProvider: new(mocks.TokenProvider),
			wantErr:       nil,
		},
		{
			name:          "nil repository",
			usersRepo:     nil,
			tokenProvider: new(mocks.TokenProvider),
			wantErr:       services.ErrUserRepositoryNil,
		},
		{
			name:          "nil repository",
			usersRepo:     new(mocks.UserRepository),
			tokenProvider: nil,
			wantErr:       services.ErrTokenProviderNil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us, err := services.NewUserService(tt.usersRepo, tt.tokenProvider)
			if tt.wantErr != nil {
				require.Nil(t, us)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, us)
		})
	}
}

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		password string

		wantErr error

		mocksSetup func(repo *mocks.UserRepository)
	}{
		{
			name:     "success",
			username: "alex123",
			email:    "alex@example.com",
			password: "strong_passwordiueh2fi",

			wantErr: nil,

			mocksSetup: func(repo *mocks.UserRepository) {
				repo.On("Create", mock.AnythingOfType("*models.User")).Once().Return(nil)
			},
		},
		{
			name:     "invalid username",
			username: "影",
			email:    "alex@example.com",
			password: "strong_passwordiueh2fi",

			wantErr: vo.ErrUsernameTooShort,
		},
		{
			name:     "invalid email",
			username: "alex123",
			email:    "invalid",
			password: "strong_passwordiueh2fi",

			wantErr: vo.ErrEmailInvalid,
		},
		{
			name:     "invalid email",
			username: "alex123",
			email:    "alex@example.com",
			password: "паролькириллицей",

			wantErr: vo.ErrPasswordInvalid,
		},
		{
			name:     "internal error",
			username: "alex123",
			email:    "alex@example.com",
			password: "strong_passwordiueh2fi",

			wantErr: services.ErrUserRegisterFailed,

			mocksSetup: func(repo *mocks.UserRepository) {
				repo.On("Create", mock.AnythingOfType("*models.User")).
					Once().
					Return(errors.New("failed to save user in the database"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.UserRepository)
			tokenProvider := new(mocks.TokenProvider)
			if tt.mocksSetup != nil {
				tt.mocksSetup(repo)
			}

			us, err := services.NewUserService(repo, tokenProvider)
			require.NoError(t, err)
			require.NotNil(t, us)

			err = us.Register(tt.username, tt.email, tt.password)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	correctUser, err := models.NewUser("alex123", "correct@example.com", "correct_pass")
	require.NoError(t, err)
	require.NotNil(t, correctUser)

	tests := []struct {
		name     string
		email    string
		password string

		wantToken string
		wantErr   error

		mocksSetup func(repo *mocks.UserRepository, tokenProvider *mocks.TokenProvider)
	}{
		{
			name:     "success",
			email:    "correct@example.com",
			password: "correct_pass",

			wantToken: "lets_pretend_this_is_a_good_vaild_token",
			wantErr:   nil,

			mocksSetup: func(repo *mocks.UserRepository, tokenProvider *mocks.TokenProvider) {
				repo.On("FindByEmail", "correct@example.com").Once().Return(correctUser, nil)

				tokenProvider.On("Generate", mock.AnythingOfType("string")).
					Once().
					Return("lets_pretend_this_is_a_good_vaild_token", nil)
			},
		},
		{
			name:     "user not found",
			email:    "notfound@example.com",
			password: "any",

			wantToken: "",
			wantErr:   services.ErrUserNotFound,

			mocksSetup: func(repo *mocks.UserRepository, tokenProvider *mocks.TokenProvider) {
				repo.On("FindByEmail", "notfound@example.com").Once().
					Return(nil, services.ErrUserRepoNotFound)
			},
		},
		{
			name:     "internal error",
			email:    "correct@example.com",
			password: "correct_pass",

			wantToken: "",
			wantErr:   services.ErrUserLoginFailed,

			mocksSetup: func(repo *mocks.UserRepository, tokenProvider *mocks.TokenProvider) {
				repo.On("FindByEmail", "correct@example.com").
					Return(nil, errors.New("db down"))
			},
		},
		{
			name:     "password mismatch",
			email:    "correct@example.com",
			password: "wrongpass",

			wantToken: "",
			wantErr:   services.ErrUserUnauthorized,

			mocksSetup: func(repo *mocks.UserRepository, tokenProvider *mocks.TokenProvider) {
				repo.On("FindByEmail", "correct@example.com").Once().Return(correctUser, nil)
			},
		},
		{
			name:     "token generation fails",
			email:    "correct@example.com",
			password: "correct_pass",

			wantToken: "",
			wantErr:   services.ErrUserLoginFailed,

			mocksSetup: func(repo *mocks.UserRepository, tokenProvider *mocks.TokenProvider) {
				repo.On("FindByEmail", "correct@example.com").Once().Return(correctUser, nil)

				tokenProvider.On("Generate", mock.AnythingOfType("string")).
					Once().
					Return("", errors.New("token generation failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.UserRepository)
			tokenProvider := new(mocks.TokenProvider)
			if tt.mocksSetup != nil {
				tt.mocksSetup(repo, tokenProvider)
			}

			us, err := services.NewUserService(repo, tokenProvider)
			require.NoError(t, err)
			require.NotNil(t, us)

			token, err := us.Login(tt.email, tt.password)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantToken, token)
		})
	}
}
