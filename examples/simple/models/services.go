package models

import "context"

// @proto.service
type UserService interface {
	// @rpc name:"GetUser" input:"GetUserRequest" output:"User"
	GetUser(ctx context.Context, req GetUserRequest) (User, error)

	// @rpc name:"CreateUser" input:"CreateUserRequest" output:"CreateUserResponse"
	CreateUser(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error)
}
