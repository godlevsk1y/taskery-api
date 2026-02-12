package user

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers"
	myMw "github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/middleware"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

type Deleter interface {
	Delete(ctx context.Context, id string, password string) error
}

type DeleteHandler struct {
	deleter  Deleter
	timeout  time.Duration
	logger   *slog.Logger
	validate *validator.Validate
}

func NewDeleteHandler(
	deleter Deleter,
	timeout time.Duration,
	logger *slog.Logger,
	validate *validator.Validate) *DeleteHandler {

	return &DeleteHandler{
		deleter:  deleter,
		timeout:  timeout,
		logger:   logger,
		validate: validate,
	}
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.User.Delete"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req, ok := handlers.DecodeAndValidate[DeleteRequest](w, r, logger, h.validate)
	if !ok {
		return
	}

	userID := r.Context().Value(myMw.UserContextKey).(string)

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	err := h.deleter.Delete(ctx, userID, req.Password)
	if err != nil {
		logger.Error("failed to delete user",
			slog.String("error", err.Error()),
			slog.String("user_id", userID),
		)

		if errors.Is(err, services.ErrUserNotFound) {
			handlers.WriteError(w, http.StatusNotFound, errors.New("user not found"))
			return
		}

		if errors.Is(err, services.ErrUserUnauthorized) {
			handlers.WriteError(w, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		handlers.WriteError(w, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	handlers.WriteJSON(w, http.StatusNoContent, nil)
}
