package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/example/petstore/generated/adapter"
	pb "github.com/example/petstore/generated/proto/v1"
	"github.com/example/petstore/models"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a client that works with Go types (not protobuf types)
	client := adapter.NewPetServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a pet using Go types - the adapter handles protobuf conversion automatically
	createReq := models.CreatePetRequest{
		Name:        "Buddy",
		Category:    models.PetCategoryDog,
		Breed:       "Golden Retriever",
		Age:         3,
		Description: "Friendly and energetic dog",
	}

	log.Println("Creating pet with Go types...")
	pet, err := client.CreatePet(ctx, createReq)
	if err != nil {
		log.Fatalf("Failed to create pet: %v", err)
	}

	log.Printf("Created pet: %+v", pet)

	// Get the pet back
	getReq := models.GetPetRequest{
		ID: pet.ID,
	}

	log.Println("Getting pet...")
	retrievedPet, err := client.GetPet(ctx, getReq)
	if err != nil {
		log.Fatalf("Failed to get pet: %v", err)
	}

	log.Printf("Retrieved pet: %+v", retrievedPet)

	// You can also use the raw protobuf client directly if needed
	pbClient := pb.NewPetServiceClient(conn)

	// But you'd have to work with protobuf types manually
	pbCreateReq := &pb.CreatePetRequest{
		Name:        "Rex",
		Category:    pb.PetCategory_PET_CATEGORY_DOG,
		Breed:       "German Shepherd",
		Age:         5,
		Description: "Guard dog",
	}

	log.Println("Creating pet with protobuf types...")
	pbPet, err := pbClient.CreatePet(ctx, pbCreateReq)
	if err != nil {
		log.Fatalf("Failed to create pet with protobuf client: %v", err)
	}

	log.Printf("Created pet with protobuf: %+v", pbPet)
}
