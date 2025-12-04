package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/example/petstore/generated/adapter"
	pb "github.com/example/petstore/generated/proto/v1"
	"github.com/example/petstore/models"
)

func main() {
	// Create a TCP listener on port 8080
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a gRPC server
	s := grpc.NewServer()

	// Create concrete service instances
	petService := models.NewPetService()
	adopterService := models.NewAdopterService()

	// Create and register adapters that bridge between our Go types and protobuf types
	petAdapter := adapter.NewPetServiceAdapter(*petService)
	adopterAdapter := adapter.NewAdopterServiceAdapter(*adopterService)

	// Register the adapters with the gRPC server
	pb.RegisterPetServiceServer(s, petAdapter)
	pb.RegisterAdopterServiceServer(s, adopterAdapter)

	// Enable gRPC reflection for easier testing with grpcurl
	reflection.Register(s)

	log.Println("Starting gRPC server on :8080")
	log.Println("The server bridges between Go structs and protobuf types automatically")
	log.Println("You can test with: grpcurl -plaintext localhost:8080 list")

	// Start the server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
