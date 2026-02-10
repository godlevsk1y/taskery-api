package user

// ========= Requests =================

type UpdateRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Username string `json:"username" validate:"omitempty"`

	Password string `json:"password" validate:"required,printascii"`
}

type DeleteRequest struct {
	Password string `json:"password" validate:"required,printascii"`
}

// ========= Responses ================

type UpdateResponse struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}
