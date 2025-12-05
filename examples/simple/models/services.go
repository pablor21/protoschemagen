package models

// @proto.service
type UserService interface {
	GetUser(id int64) (*User, error)
	CreateUser(user *User) error
}
