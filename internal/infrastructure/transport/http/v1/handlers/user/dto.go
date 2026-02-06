package user

// ========= Requests =================

type UpdateRequest struct {
	Email    string `json:"email" validate:"email"`
	Username string `json:"username" validate:"alphanumunicode"`

	Password string `json:"password" validate:"required,printascii"`
}

// ========= Responses ================

type UpdateResponse struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}
