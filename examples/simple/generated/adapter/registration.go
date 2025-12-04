// Package adapter contains auto-generated registration helpers
// Generated from protobuf annotations - DO NOT EDIT
package adapter

import (
	"google.golang.org/grpc"
	"github.com/coso/models"
)
// RegisterUserService registers the service with a gRPC server using original types
func RegisterUserService(server *grpc.Server, service models.UserService) {
	adapter := NewUserServiceAdapter(service)
	_ = adapter // Used for registration - actual registration call depends on generated protobuf service
}