package user

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
	const op = "handlers.user.Register"

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req RegisterRequest
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

	if err := h.validate.Struct(req); err != nil {
		logger.Error("failed to validate request body", slog.String("error", err.Error()))
		handlers.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.registrar.Register(ctx, req.Username, req.Email, req.Password)
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
