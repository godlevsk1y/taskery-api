package task

import "time"

// ========= Requests =================

type CreateRequest struct {
	Title       string     `json:"title" validate:"required"`
	Description string     `json:"description" validate:"required"`
	Deadline    *time.Time `json:"deadline"`
}

type DeleteRequest struct {
	TaskID string `json:"task_id" validate:"required"`
}

// ========= Responses ================

type CreateResponse struct {
	TaskID string `json:"task_id"`
}
