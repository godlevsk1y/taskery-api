package user

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers"
	"github.com/go-chi/chi/v5/middleware"
)

// Registrar is an interface that wraps Register function for registration of new users
type Registrar interface {
	Register(ctx context.Context, username, email, password string) error
}

func Register(ctx context.Context, reg Registrar, logger *slog.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.Register"

		logger = logger.With(
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
			handlers.WriteError(w, http.StatusBadRequest, errors.New("request body is empty"))
			return
		}

		// TODO: to be continued (we have parsed request before this line)
	}
}
