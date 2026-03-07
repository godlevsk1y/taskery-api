package task

import "time"

// ========= Requests =================

type CreateRequest struct {
	Title       string     `json:"title" validate:"required"`
	Description string     `json:"description" validate:"required"`
	Deadline    *time.Time `json:"deadline"`
}

type UpdateRequest struct {
	TaskID      string     `json:"task_id" validate:"required"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Deadline    *time.Time `json:"deadline"`
}

type DeleteRequest struct {
	TaskID string `json:"task_id" validate:"required"`
}

type CompleteRequest struct {
	TaskID string `json:"task_id" validate:"required"`
}

type RemoveDeadlineRequest struct {
	TaskID string `json:"task_id" validate:"required"`
}

type ReopenRequest struct {
	TaskID string `json:"task_id" validate:"required"`
}

// ========= Responses ================

type CreateResponse struct {
	TaskID string `json:"task_id"`
}

type UpdateResponse struct {
	TaskID string `json:"task_id"`
}
