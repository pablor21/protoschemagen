package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/coso/generated/adapter"
	v1 "github.com/coso/generated/proto/v1"
	"github.com/coso/models"
)

func main() {
	// Create a gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// Instantiate the service implementation
	service := &UserServiceImpl{}

	// Create the adapter that bridges original types to protobuf
	adapterService := adapter.NewUserServiceAdapter(service)

	// Register the service with the gRPC server
	v1.RegisterUserServiceServer(s, adapterService)

	// Enable gRPC reflection for easier debugging
	reflection.Register(s)

	log.Println("Starting gRPC server on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

type UserServiceImpl struct{}

func (s *UserServiceImpl) GetUser(req int64) (*models.User, error) {
	// Implement your logic to get a user by ID
	log.Printf("GetUser called with ID: %d", req)
	return &models.User{
		ID:       req,
		Username: "example_user",
		Email:    "example@example.com",
	}, nil
}

func (s *UserServiceImpl) CreateUser(req *models.User) error {
	// Implement your logic to create a user
	log.Printf("CreateUser called with user: %+v", req)
	return nil
}

// func (s *UserServiceImpl) GetUser(ctx context.Context, req models.GetUserRequest) (models.User, error) {
// 	// Implement your logic to get a user by ID
// 	log.Printf("GetUser called with ID: %d", req.ID)
// 	return models.User{
// 		ID:       req.ID,
// 		Username: "example_user",
// 		Email:    "example@example.com",
// 	}, nil
// }

// func (s *UserServiceImpl) CreateUser(ctx context.Context, req models.CreateUserRequest) (models.CreateUserResponse, error) {
// 	// Implement your logic to create a user
// 	log.Printf("CreateUser called with user: %+v", req.User)
// 	return models.CreateUserResponse{
// 		Success: true,
// 		Message: "User created successfully",
// 	}, nil
// }
