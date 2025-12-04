// Package adapter contains auto-generated gRPC service adapter
// Generated from protobuf annotations - DO NOT EDIT
package adapter

import (
	"github.com/coso/models"
	"context"
	pb "github.com/coso/generated/proto/v1"
)

// UserServiceAdapter adapts models.UserService to gRPC pb.UserServiceServer
type UserServiceAdapter struct {
	pb.UnimplementedUserServiceServer
	service models.UserService
}

// NewUserServiceAdapter creates a new UserServiceAdapter
func NewUserServiceAdapter(service models.UserService) *UserServiceAdapter {
	return &UserServiceAdapter{
		service: service,
	}
}
// GetUser implements unary RPC
func (a *UserServiceAdapter) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	// Convert request
	goReq := GetUserRequestFromProto(req)
	
	// Call service method
	result, err := a.service.GetUser(ctx, goReq)
	if err != nil {
		return nil, err
	}
	
	// Convert and return response
	return UserToProto(result), nil
}
// CreateUser implements unary RPC
func (a *UserServiceAdapter) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Convert request
	goReq := CreateUserRequestFromProto(req)
	
	// Call service method
	result, err := a.service.CreateUser(ctx, goReq)
	if err != nil {
		return nil, err
	}
	
	// Convert and return response
	return CreateUserResponseToProto(result), nil
}