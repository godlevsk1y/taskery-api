package auth

import (
	"context"
	"errors"
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

	req, ok := handlers.DecodeAndValidate[LoginRequest](w, r, l.logger, l.validate)
	if !ok {
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
