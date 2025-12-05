// Package adapter contains auto-generated gRPC client
// Generated from protobuf annotations - DO NOT EDIT
package adapter

import (
	"github.com/coso/models"
	"context"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
func (c *UserServiceClient) GetUser(req int64) (*models.User, error) {
	// Convert primitive to wrapper
	protoReq := &wrapperspb.Int64Value{Value: req}

	// Call gRPC method
	protoResp, err := c.client.GetUser(context.Background(), protoReq)
	if err != nil {
		return nil, err
	}
	// Convert protobuf response back to Go types
	result := UserFromProto(protoResp)
	return &result, nil
}
// CreateUser calls the gRPC CreateUser method using Go types
func (c *UserServiceClient) CreateUser(req *models.User) error {
	// Convert Go request to protobuf
	protoReq := UserToProto(*req)

	// Call gRPC method
	_, err := c.client.CreateUser(context.Background(), protoReq)
	if err != nil {
		return err
	}
	return nil
}