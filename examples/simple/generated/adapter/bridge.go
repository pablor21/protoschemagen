// Package adapter contains auto-generated service bridge
// Generated from protobuf annotations - DO NOT EDIT
package adapter

import (
	"github.com/coso/models"
)
// UserServiceBridge implements the adapter interface by directly using the original service
type UserServiceBridge struct {
	originalService models.UserService
}

// NewUserServiceBridge creates a bridge that directly wraps the original service
func NewUserServiceBridge(service models.UserService) models.UserService {
	return &UserServiceBridge{
		originalService: service,
	}
}
// GetUser implements the adapter interface by calling the original service
func (b *UserServiceBridge) GetUser(req int64) (*models.User, error) {
	return b.originalService.GetUser(req)
}
// CreateUser implements the adapter interface by calling the original service
func (b *UserServiceBridge) CreateUser(req *models.User) error {
	return b.originalService.CreateUser(req)
}