package v1

import (
	"context"
	"log/slog"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/models"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/auth"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/task"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/user"
	myMw "github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/middleware"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

type UserService interface {
	Register(ctx context.Context, username, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
	ChangeUsername(ctx context.Context, id, newUsername, password string) error
	ChangeEmail(ctx context.Context, id, newEmail, password string) error
	ChangePassword(ctx context.Context, id, old, new string) error
	Delete(ctx context.Context, id, password string) error
}

type TaskService interface {
	Create(ctx context.Context, cmd services.CreateTaskCommand) (string, error)
	Update(ctx context.Context, id string, ownerID string, cmd services.UpdateTaskCommand) error
	RemoveDeadline(ctx context.Context, id string, ownerID string) error
	Complete(ctx context.Context, id string, ownerID string) error
	Reopen(ctx context.Context, id string, ownerID string) error
	FindByOwner(ctx context.Context, ownerID string) ([]*models.Task, error)
	Delete(ctx context.Context, id string, ownerID string) error
}

type TokenProvider interface {
	Generate(userID string) (string, error)
	Validate(token string) (string, error)
}

type RouterOptions struct {
	UserService UserService
	TaskService TaskService

	Logger        *slog.Logger
	TokenProvider TokenProvider
	Validator     *validator.Validate

	Timeout time.Duration
}

func NewRouter(opts RouterOptions) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID, middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Method("POST", "/register", auth.NewRegisterHandler(
				opts.UserService,
				opts.Timeout,
				opts.Logger,
				opts.Validator,
			))
			r.Method("POST", "/login", auth.NewLoginHandler(
				opts.UserService,
				opts.Timeout,
				opts.Logger,
				opts.Validator,
			))
		})

		r.Group(func(r chi.Router) {
			r.Use(myMw.JWTAuth(opts.TokenProvider, opts.Logger))
			r.Method("PATCH", "/users", user.NewUpdateHandler(
				opts.UserService,
				opts.Timeout,
				opts.Logger,
				opts.Validator,
			))
			r.Method("DELETE", "/users", user.NewDeleteHandler(
				opts.UserService,
				opts.Timeout,
				opts.Logger,
				opts.Validator,
			))
		})

		r.Group(func(r chi.Router) {
			r.Use(myMw.JWTAuth(opts.TokenProvider, opts.Logger))

			r.Method("POST", "/tasks", task.NewCreateHandler(
				opts.TaskService,
				opts.Timeout,
				opts.Logger,
				opts.Validator,
			))

			r.Method("GET", "/tasks", task.NewFindByOwnerHandler(
				opts.TaskService,
				opts.Timeout,
				opts.Logger,
				opts.Validator,
			))

			r.Method("PATCH", "/tasks", task.NewUpdateHandler(
				opts.TaskService,
				opts.Timeout,
				opts.Logger,
				opts.Validator,
			))

			r.Method("DELETE", "/tasks", task.NewDeleteHandler(
				opts.TaskService,
				opts.Timeout,
				opts.Logger,
				opts.Validator,
			))

			r.Method("PATCH", "/tasks/complete", task.NewCompleteHandler(
				opts.TaskService,
				opts.Timeout,
				opts.Logger,
				opts.Validator,
			))

			r.Method("PATCH", "/tasks/reopen", task.NewReopenHandler(
				opts.TaskService,
				opts.Timeout,
				opts.Logger,
				opts.Validator,
			))

			r.Method("PATCH", "/tasks/remove-deadline", task.NewRemoveDeadlineHandler(
				opts.TaskService,
				opts.Timeout,
				opts.Logger,
				opts.Validator,
			))
		})
	})

	return r
}
