package models

import "context"

// @proto.service
type UserService interface {
	GetUser(ctx context.Context, req GetUserRequest) (User, error)
	CreateUser(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error)
}
