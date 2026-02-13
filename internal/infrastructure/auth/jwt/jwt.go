package jwt

import (
	"fmt"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/golang-jwt/jwt/v5"
)

// Provider is responsible for issuing and validating JSON Web Tokens (JWT).
//
// It encapsulates the signing secret, token time-to-live (TTL),
// and the issuer identifier used in JWT claims.
// The provider is typically used by authentication or authorization
// layers to generate access tokens and verify their validity.
type Provider struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

// NewProvider a new instance of jwt.Provider
func NewProvider(secret []byte, ttl time.Duration, issuer string) *Provider {
	return &Provider{secret: secret, ttl: ttl, issuer: issuer}
}

type Claims struct {
	jwt.RegisteredClaims
}

// Generate creates a signed JWT token for the given userID.
// The token contains standard claims: "sub" (subject) set to userID,
// "iat" (issued at) set to the current Unix time, "exp" (expiration) set
// to the current time plus the provider's TTL, and "iss" (issuer) set
// to the provider's issuer string.
//
// Generate returns the signed JWT as a string and any error encountered
// while signing the token.
func (p *Provider) Generate(userID string) (string, error) {
	const op = "jwt.Provider.Generate"

	claims := &Claims{
		jwt.RegisteredClaims{
			Issuer:    p.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(p.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(p.secret)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return signed, nil
}

// Validate checks a JWT token and returns its "sub" claim if valid. It returns user's ID and an error
func (p *Provider) Validate(token string) (string, error) {
	const op = "jwt.Provider.Validate"

	claims := &Claims{}

	parsedToken, err := jwt.ParseWithClaims(
		token,
		claims,
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%s: unexpected signing method: %v", op, token.Header["alg"])
			}

			return p.secret, nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("%s: parse token: %w", op, err)
	}

	if !parsedToken.Valid {
		return "", fmt.Errorf("%s: token is invalid", op)
	}

	if claims.Issuer != p.issuer {
		return "", fmt.Errorf("%s: invalid issuer", op)
	}

	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return "", fmt.Errorf("%s: expired token", op)
	}

	if claims.Subject == "" {
		return "", fmt.Errorf("%s: subject is empty", op)
	}

	return claims.Subject, nil
}

var _ services.TokenProvider = (*Provider)(nil)
