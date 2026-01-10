package models_test

import (
	"testing"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/models"
	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/vo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		email       string
		password    string
		expectedErr error
	}{
		{
			name:        "valid inputs",
			username:    "john_doe",
			email:       "john@example.com",
			password:    "Str0ngP@ssw0rd!",
			expectedErr: nil,
		},
		{
			name:        "invalid username",
			username:    "山",
			email:       "john@example.com",
			password:    "Str0ngP@ssw0rd!",
			expectedErr: vo.ErrUsernameTooShort,
		},
		{
			name:        "invalid email",
			username:    "john_doe",
			email:       "not-an-email",
			password:    "Str0ngP@ssw0rd!",
			expectedErr: vo.ErrEmailInvalid,
		},
		{
			name:        "invalid password",
			username:    "john_doe",
			email:       "john@example.com",
			password:    "short",
			expectedErr: vo.ErrPasswordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := models.NewUser(tt.username, tt.email, tt.password)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, u)
			require.NotEqual(t, uuid.Nil, u.ID())

			// Check that VOs are populated correctly
			require.Equal(t, tt.username, u.Username().String())
			require.Equal(t, tt.email, u.Email().String())

			// Verify password is hashed and can be verified
			require.NoError(t, u.PasswordHash().Verify(tt.password))
		})
	}
}

func TestNewUserFromDB(t *testing.T) {
	validID := uuid.New().String()

	tests := []struct {
		name        string
		id          string
		username    string
		email       string
		password    string
		expectedErr error
	}{
		{
			name:        "valid inputs with id",
			id:          validID,
			username:    "john_doe",
			email:       "john@example.com",
			password:    "Str0ngP@ssw0rd!",
			expectedErr: nil,
		},
		{
			name:        "invalid id",
			id:          "not-a-uuid",
			username:    "john_doe",
			email:       "john@example.com",
			password:    "Str0ngP@ssw0rd!",
			expectedErr: models.ErrUserIDInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := models.NewUserFromDB(models.UserFromDBParams{
				ID:           tt.id,
				Username:     tt.username,
				Email:        tt.email,
				PasswordHash: tt.password,
			})
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, u)

			parsed, err := uuid.Parse(tt.id)
			require.NoError(t, err)
			require.Equal(t, parsed, u.ID())

			require.Equal(t, u.PasswordHash().String(), tt.password)
		})
	}
}

func TestChangeUsername(t *testing.T) {
	u, err := models.NewUser("john_doe", "john@example.com", "Str0ngP@ssw0rd!")
	require.NoError(t, err)

	tests := []struct {
		name    string
		newVal  string
		wantErr bool
	}{
		{
			name:    "valid username",
			newVal:  "new_john",
			wantErr: false,
		},
		{
			name:    "invalid username",
			newVal:  "山",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.ChangeUsername(tt.newVal)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.newVal, u.Username().String())
		})
	}
}

func TestChangeEmail(t *testing.T) {
	u, err := models.NewUser("john_doe", "john@example.com", "Str0ngP@ssw0rd!")
	require.NoError(t, err)

	tests := []struct {
		name    string
		newVal  string
		wantErr bool
	}{
		{
			name:    "valid email",
			newVal:  "newjohn@example.com",
			wantErr: false,
		},
		{
			name:    "invalid email",
			newVal:  "not-an-email",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.ChangeEmail(tt.newVal)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.newVal, u.Email().String())
		})
	}
}

func TestChangePassword(t *testing.T) {
	u, err := models.NewUser("john_doe", "john@example.com", "Str0ngP@ssw0rd!")
	require.NoError(t, err)

	tests := []struct {
		name    string
		oldPass string
		newPass string
		wantErr bool
	}{
		{
			name:    "valid change",
			oldPass: "Str0ngP@ssw0rd!",
			newPass: "An0ther$trongPass!",
			wantErr: false,
		},
		{
			name:    "wrong old password",
			oldPass: "wrong",
			newPass: "An0ther$trongPass!",
			wantErr: true,
		},
		{
			name:    "invalid new password",
			oldPass: "Str0ngP@ssw0rd!",
			newPass: "short",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.ChangePassword(tt.oldPass, tt.newPass)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			// Verify the new password works
			require.NoError(t, u.PasswordHash().Verify(tt.newPass))
			// And old no longer works
			require.Error(t, u.PasswordHash().Verify(tt.oldPass))
		})
	}
}
