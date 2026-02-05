package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/cyberbrain-dev/taskery-api/pkg/errorsx"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

type Authenticator interface {
	Login(ctx context.Context, email, password string) (string, error)
}

type LoginHandler struct {
	authenticator Authenticator
	timeout       time.Duration
	logger        *slog.Logger
	validate      *validator.Validate
}

func NewLoginHandler(
	authenticator Authenticator,
	timeout time.Duration,
	logger *slog.Logger,
	validate *validator.Validate) *LoginHandler {
	return &LoginHandler{
		authenticator: authenticator,
		timeout:       timeout,
		logger:        logger,
		validate:      validate,
	}
}

func (l *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.auth.Login"

	ctx, cancel := context.WithTimeout(r.Context(), l.timeout)
	defer cancel()

	logger := l.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if errors.Is(err, io.EOF) {
		logger.Error("request body is empty")
		handlers.WriteError(w, http.StatusBadRequest, errors.New("request body is empty"))
		return
	}
	if err != nil {
		logger.Error("failed to decode request body")
		handlers.WriteError(w, http.StatusBadRequest, errors.New("failed to decode request body"))
		return
	}

	if err := l.validate.Struct(req); err != nil {
		logger.Error("failed to validate request body", slog.String("error", err.Error()))
		handlers.WriteError(w, http.StatusBadRequest, err)
		return
	}

	token, err := l.authenticator.Login(ctx, req.Email, req.Password)
	if err != nil {
		logger.Error("failed to login", slog.String("error", err.Error()))

		if errorsx.IsAny(err, services.ErrUserNotFound, services.ErrUserUnauthorized) {
			handlers.WriteError(w, http.StatusUnauthorized, errors.New("invalid credentials"))
			return
		}

		handlers.WriteError(w, http.StatusInternalServerError, err)
	}

	handlers.WriteJSON(w, http.StatusOK, LoginResponse{
		Token: token,
	})
}
