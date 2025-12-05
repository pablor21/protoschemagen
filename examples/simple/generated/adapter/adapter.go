// Package adapter contains auto-generated gRPC service adapter
// Generated from protobuf annotations - DO NOT EDIT
package adapter

import (
	"context"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"google.golang.org/protobuf/types/known/emptypb"
	"github.com/coso/models"
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
func (a *UserServiceAdapter) GetUser(ctx context.Context, req *wrapperspb.Int64Value) (*pb.User, error) {
	// Convert request from protobuf to original Go type
	goReq := req.Value  // Extract int64 from wrapper
	
	// Call service method with original signature
	result, err := a.service.GetUser(goReq)
	if err != nil {
		return nil, err
	}
	
	// Convert result to protobuf type
	return UserToProto(*result), nil
}
// CreateUser implements unary RPC
func (a *UserServiceAdapter) CreateUser(ctx context.Context, req *pb.User) (*emptypb.Empty, error) {
	// Convert request from protobuf to original Go type
	goReqValue := UserFromProto(req)
	goReq := &goReqValue
	
	// Call service method with original signature
	err := a.service.CreateUser(goReq)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}