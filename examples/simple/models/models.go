package models

// @proto.message
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// @proto.message
type GetUserRequest struct {
	ID int64 `json:"id"`
}

// @proto.message
type CreateUserRequest struct {
	User *User `json:"user"`
}

// @proto.message
type CreateUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
