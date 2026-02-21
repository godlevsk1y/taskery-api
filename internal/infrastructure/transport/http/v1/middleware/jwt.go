package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers"
	"github.com/go-chi/chi/v5/middleware"
)

type contextKey string

// UserContextKey is the string literal used in HTTP-request's context as key for user ID
const UserContextKey contextKey = "userID"

// JWTValidator wraps a method for parsing and validating JWT token.
type JWTValidator interface {
	// Validate gets a JWT token and returns user ID from this token and an error, if occurred.
	Validate(token string) (string, error)
}

// TODO: Write docs for this func
func JWTAuth(validator JWTValidator, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.JWT"

			logger := logger.With(
				slog.String("op", op),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Error("no token provided")

				handlers.WriteError(w, http.StatusUnauthorized, errors.New("unauthorized"))
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				logger.Error("invalid Authorization format")

				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := validator.Validate(tokenString)
			if err != nil {
				logger.Error("failed to validate token", slog.String("error", err.Error()))

				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
