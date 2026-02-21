package task

import "time"

// ========= Requests =================

type CreateRequest struct {
	Title       string     `json:"title" validate:"required"`
	Description string     `json:"description" validate:"required"`
	Deadline    *time.Time `json:"deadline"`
}

// ========= Responses ================

// TODO: REFACTOR THE CREATE TASK RESPONSE
type CreateResponse struct {
	Title string `json:"title"`
}
