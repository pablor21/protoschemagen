// Package adapter contains auto-generated gRPC client
// Generated from protobuf annotations - DO NOT EDIT
package adapter

import (
	"github.com/coso/models"
	"context"
	pb "github.com/coso/generated/proto/v1"
	"google.golang.org/grpc"
)
// UserServiceClient wraps the gRPC client and provides Go types interface
type UserServiceClient struct {
	client pb.UserServiceClient
}

// NewUserServiceClient creates a new client for the service
func NewUserServiceClient(conn *grpc.ClientConn) *UserServiceClient {
	return &UserServiceClient{
		client: pb.NewUserServiceClient(conn),
	}
}
// GetUser calls the gRPC GetUser method using Go types
func (c *UserServiceClient) GetUser(ctx context.Context, req models.GetUserRequest) (models.User, error) {
	// Convert Go request to protobuf
	protoReq := GetUserRequestToProto(req)

	// Call gRPC method
	protoResp, err := c.client.GetUser(ctx, protoReq)
	if err != nil {
		return models.User{}, err
	}

	// Convert protobuf response back to Go types
	return UserFromProto(protoResp), nil
}
// CreateUser calls the gRPC CreateUser method using Go types
func (c *UserServiceClient) CreateUser(ctx context.Context, req models.CreateUserRequest) (models.CreateUserResponse, error) {
	// Convert Go request to protobuf
	protoReq := CreateUserRequestToProto(req)

	// Call gRPC method
	protoResp, err := c.client.CreateUser(ctx, protoReq)
	if err != nil {
		return models.CreateUserResponse{}, err
	}

	// Convert protobuf response back to Go types
	return CreateUserResponseFromProto(protoResp), nil
}