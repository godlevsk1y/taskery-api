package jwt

import (
	"time"

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

func (p Provider) Generate(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(p.ttl).Unix(),
		"iss": p.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(p.secret)
	if err != nil {
		return "", err
	}

	return signed, nil
}
