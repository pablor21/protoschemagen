package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/example/proto/starwars/generated/adapter"
	pb "github.com/example/proto/starwars/generated/proto/v1"
	"github.com/example/proto/starwars/models"
)

func main() {
	// Create the in-memory service (slice-based with context)
	service := NewInMemoryStarWarsService()

	// Create a bridge that converts slice interface to channel interface for the adapter
	bridgedService := adapter.NewStarWarsServiceBridge(service)

	// Create the gRPC adapter with the bridged service
	grpcAdapter := adapter.NewStarWarsServiceAdapter(bridgedService)

	// Create gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// Register the service
	pb.RegisterStarWarsServiceServer(s, grpcAdapter)

	log.Println("🚀 Star Wars gRPC Server starting...")
	log.Println("📡 Server listening on :50051")
	log.Println("🌟 Service includes:")

	stats := service.GetStats()
	log.Printf("   - %v humans", stats["total_humans"])
	log.Printf("   - %v droids", stats["total_droids"])

	// Test the adapter functionality before starting the server
	log.Println("\n🧪 Testing adapter functionality:")

	// Test GetHuman
	ctx := context.Background()

	// Create protobuf request (this would normally come from a client)
	protoReq := &pb.GetHumanRequest{
		Id: "luke-skywalker",
	}

	// Call the adapter method
	protoResp, err := grpcAdapter.GetHuman(ctx, protoReq)
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		log.Printf("✅ GetHuman successful: %s (Height: %.2f)", protoResp.Name, protoResp.Height)
	}

	// Test GetDroid
	protoDroidReq := &pb.GetDroidRequest{
		Id: "c-3po",
	}

	protoDroidResp, err := grpcAdapter.GetDroid(ctx, protoDroidReq)
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		log.Printf("✅ GetDroid successful: %s (Function: %s)", protoDroidResp.Name, protoDroidResp.PrimaryFunction)
	}

	// Demonstrate type conversion
	log.Println("\n🔄 Type conversion examples:")

	// Original model -> Protobuf
	originalHuman := models.Human{
		ID:     "test-human",
		Name:   "Test Character",
		Height: 1.75,
	}
	protoHuman := adapter.HumanToProto(originalHuman)
	log.Printf("📤 Original -> Proto: %s (ID: %s)", protoHuman.Name, protoHuman.Id)

	// Protobuf -> Original model
	convertedBack := adapter.HumanFromProto(protoHuman)
	log.Printf("📥 Proto -> Original: %s (ID: %s)", convertedBack.Name, convertedBack.ID)

	// Test enum conversion
	episodes := []models.Episode{models.NEWHOPE, models.EMPIRE}
	protoEpisodes := adapter.EpisodeSliceToProto(episodes)
	log.Printf("📺 Episodes converted: %d episodes to protobuf", len(protoEpisodes))

	log.Println("\n🚀 Starting gRPC server...")

	// Start the server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
