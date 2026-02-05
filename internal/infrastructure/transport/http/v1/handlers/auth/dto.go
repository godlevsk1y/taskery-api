package auth

// ========= Requests =================

type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,printascii"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,printascii"`
}

// ========= Responses ================

type RegisterResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
