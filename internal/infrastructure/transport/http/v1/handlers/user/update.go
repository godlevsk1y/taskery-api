package user

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

type Updater interface {
	ChangeUsername(ctx context.Context, id, newUsername, password string) error
	ChangeEmail(ctx context.Context, id, newEmail, password string) error
}

type UpdateHandler struct {
	updater  Updater
	timeout  time.Duration
	logger   *slog.Logger
	validate *validator.Validate
}

func NewUpdateHandler(
	updater Updater,
	timeout time.Duration,
	logger *slog.Logger,
	validate *validator.Validate) *UpdateHandler {

	return &UpdateHandler{
		updater:  updater,
		timeout:  timeout,
		logger:   logger,
		validate: validate,
	}
}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.User.Update"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req, ok := handlers.DecodeAndValidate[UpdateRequest](w, r, logger, h.validate)
	if !ok {
		return
	}

	if req.Username == "" && req.Email == "" {
		handlers.WriteError(w, http.StatusBadRequest, errors.New("username or email required"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	// ! TODO: make a separate function and type for user_id (func that takes context and return string and ok)
	userID := r.Context().Value("userID").(string)
	if req.Username != "" {
		err := h.updater.ChangeUsername(ctx, userID, req.Username, req.Password)
		if err != nil {
			logger.Error("failed to change username", slog.String("error", err.Error()))

			if errorsx.IsAny(err, services.ErrUserNotFound) {
				handlers.WriteError(w, http.StatusNotFound, errors.New("user not found"))
				return
			}

			if errorsx.IsAny(err, services.ErrUserUnauthorized) {
				handlers.WriteError(w, http.StatusUnauthorized, errors.New("unauthorized"))
				return
			}

			if errorsx.IsAny(err, services.ErrUserChangeUsernameFailed) {
				handlers.WriteError(w, http.StatusInternalServerError, errors.New("username change failed"))
				return
			}

			handlers.WriteError(w, http.StatusBadRequest, err)
			return
		}
	}

	if req.Email != "" {
		err := h.updater.ChangeEmail(ctx, userID, req.Email, req.Password)
		if err != nil {
			logger.Error("failed to change email", slog.String("error", err.Error()))

			if errorsx.IsAny(err, services.ErrUserNotFound) {
				handlers.WriteError(w, http.StatusNotFound, errors.New("user not found"))
				return
			}

			if errorsx.IsAny(err, services.ErrUserUnauthorized) {
				handlers.WriteError(w, http.StatusUnauthorized, errors.New("password incorrect"))
				return
			}

			if errorsx.IsAny(err, services.ErrUserEmailAlreadyTaken) {
				handlers.WriteError(w, http.StatusConflict, errors.New("email already taken"))
				return
			}

			if errorsx.IsAny(err, services.ErrUserChangeEmailFailed) {
				handlers.WriteError(w, http.StatusInternalServerError, errors.New("email change failed"))
				return
			}

			handlers.WriteError(w, http.StatusBadRequest, err)
			return
		}
	}

	handlers.WriteJSON(w, http.StatusOK, UpdateResponse{
		Username: req.Username,
		Email:    req.Email,
	})
}
