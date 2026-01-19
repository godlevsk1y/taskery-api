package jwt_test

import (
	"testing"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/auth/jwt"
	"github.com/stretchr/testify/require"
)

func TestProvider_GenerateAndValidate_OK(t *testing.T) {
	p := jwt.NewProvider(
		[]byte("test-secret"),
		time.Minute,
		"test-issuer",
	)

	userID := "test-user-123"

	token, err := p.Generate(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	got, err := p.Validate(token)
	require.NoError(t, err)
	require.Equal(t, userID, got)
}

func TestProvider_GenerateAndValidate_Expired(t *testing.T) {
	p := jwt.NewProvider(
		[]byte("test-secret"),
		10*time.Millisecond,
		"test-issuer",
	)

	userID := "test-user-123"

	token, err := p.Generate(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	time.Sleep(20 * time.Millisecond)

	got, err := p.Validate(token)
	require.Error(t, err)
	require.Empty(t, got)
}

func TestProvider_GenerateAndValidate_WrongIssuer(t *testing.T) {
	p := jwt.NewProvider(
		[]byte("test-secret"),
		time.Minute,
		"test-issuer",
	)

	userID := "test-user-123"

	token, err := p.Generate(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	other := jwt.NewProvider(
		[]byte("test-secret"),
		time.Minute,
		"another-issuer",
	)

	got, err := other.Validate(token)
	require.Error(t, err)
	require.Empty(t, got)
}

func TestProvider_GenerateAndValidate_WrongSecret(t *testing.T) {
	p := jwt.NewProvider(
		[]byte("test-secret"),
		time.Minute,
		"test-issuer",
	)

	userID := "test-user-123"

	token, err := p.Generate(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	other := jwt.NewProvider(
		[]byte("another-secret"),
		time.Minute,
		"test-issuer",
	)

	got, err := other.Validate(token)
	require.Error(t, err)
	require.Empty(t, got)
}

func TestProvider_GenerateAndValidate_InvalidToken(t *testing.T) {
	p := jwt.NewProvider(
		[]byte("test-secret"),
		time.Minute,
		"test-issuer",
	)

	got, err := p.Validate("not a token")
	require.Error(t, err)
	require.Empty(t, got)
}
