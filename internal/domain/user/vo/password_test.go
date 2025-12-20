package vo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/vo"
)

func TestNewPassword(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{name: "empty", input: "", expectedErr: vo.ErrPasswordEmpty},
		{name: "too short", input: mkString('a', vo.PasswordMinLength-1), expectedErr: vo.ErrPasswordTooShort},
		{name: "too long", input: mkString('a', vo.PasswordMaxLength+1), expectedErr: vo.ErrPasswordTooLong},
		{name: "non-ascii invalid", input: "пароль123", expectedErr: vo.ErrPasswordInvalid},
		{name: "valid ascii min length", input: mkString('a', vo.PasswordMinLength), expectedErr: nil},
		{name: "valid ascii typical", input: "Password123!", expectedErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := vo.NewPassword(tt.input)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			// Verify should accept the original input
			require.True(t, p.Verify(tt.input))
			// Value should be a non-empty hash string different from raw
			require.NotEmpty(t, p.Value())
			require.NotEqual(t, tt.input, p.Value())
		})
	}
}

func TestNewPasswordFromHashAndVerify(t *testing.T) {
	// create a new password and then reconstruct from its hash
	raw := "Password123!"
	p1, err := vo.NewPassword(raw)
	require.NoError(t, err)

	// ensure Verify works on the original
	require.True(t, p1.Verify(raw))

	// reconstruct using Value() as hash
	p2 := vo.NewPasswordFromHash([]byte(p1.Value()))

	// verify that the reconstructed password verifies the same raw
	require.True(t, p2.Verify(raw))

	// Value should be equal to the original hash when created from hash
	require.Equal(t, p1.Value(), p2.Value())
}

func mkString(ch rune, n int) string {
	b := make([]rune, n)
	for i := range n {
		b[i] = ch
	}
	return string(b)
}
