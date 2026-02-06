package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/cyberbrain-dev/taskery-api/pkg/errorsx"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

// Registrar is an interface that wraps Register function for registration of new users
type Registrar interface {
	Register(ctx context.Context, username, email, password string) error
}

type RegisterHandler struct {
	registrar Registrar
	timeout   time.Duration
	logger    *slog.Logger
	validate  *validator.Validate
}

func NewRegisterHandler(
	registrar Registrar,
	timeout time.Duration,
	logger *slog.Logger,
	validate *validator.Validate) *RegisterHandler {
	return &RegisterHandler{
		registrar: registrar,
		timeout:   timeout,
		logger:    logger,
		validate:  validate,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.auth.Register"

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req, ok := handlers.DecodeAndValidate[RegisterRequest](w, r, h.logger, h.validate)
	if !ok {
		return
	}

	err := h.registrar.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		logger.Error("failed to register user", slog.String("error", err.Error()))

		if errorsx.IsAny(err, services.ErrUserExists) {
			handlers.WriteError(w, http.StatusConflict, err)
			return
		}

		handlers.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, RegisterResponse{
		Username: req.Username,
		Email:    req.Email,
	})
}
