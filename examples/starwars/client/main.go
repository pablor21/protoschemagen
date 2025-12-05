package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/example/proto/starwars/generated/adapter"
// 	"github.com/example/proto/starwars/models"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
// )

// func main() {
// 	fmt.Println("🚀 Star Wars gRPC Client Test")
// 	fmt.Println("=" + string(make([]rune, 50)))

// 	// Connect to the server
// 	fmt.Println("📡 Connecting to server at localhost:50051...")
// 	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatalf("Failed to connect: %v", err)
// 	}
// 	defer conn.Close()

// 	// Create client using the adapter
// 	client := adapter.NewStarWarsServiceClient(conn)
// 	fmt.Println("✅ Client connected successfully")

// 	// Test GetHuman
// 	fmt.Println("\n🔍 Testing GetHuman...")
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	humanReq := models.GetHumanRequest{
// 		ID: "luke-skywalker",
// 	}

// 	human, err := client.GetHuman(ctx, humanReq)
// 	if err != nil {
// 		log.Printf("❌ GetHuman failed: %v", err)
// 	} else {
// 		fmt.Printf("✅ Got human: %s (%s)\n", human.Name, human.ID)
// 		fmt.Printf("   📏 Height: %.2fm\n", human.Height)
// 		fmt.Printf("   👥 Friends: %d\n", len(human.Friends))
// 		fmt.Printf("   🚀 Starships: %v\n", human.Starships)
// 	}

// 	// Test GetDroid
// 	fmt.Println("\n🤖 Testing GetDroid...")
// 	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel2()

// 	droidReq := models.GetDroidRequest{
// 		ID: "c-3po",
// 	}

// 	droid, err := client.GetDroid(ctx2, droidReq)
// 	if err != nil {
// 		log.Printf("❌ GetDroid failed: %v", err)
// 	} else {
// 		fmt.Printf("✅ Got droid: %s (%s)\n", droid.Name, droid.ID)
// 		fmt.Printf("   🔧 Function: %s\n", droid.PrimaryFunction)
// 	}

// 	// Test StreamHumans
// 	fmt.Println("\n📡 Testing StreamHumans...")
// 	ctx3, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel3()

// 	streamReq := models.StreamHumansRequest{
// 		Offset: 0,
// 		Limit:  10,
// 	}

// 	fmt.Printf("   📋 Request: Offset=%d, Limit=%d\n", streamReq.Offset, streamReq.Limit)

// 	humanChan, err := client.StreamHumans(ctx3, streamReq)
// 	if err != nil {
// 		log.Printf("❌ StreamHumans failed: %v", err)
// 	} else {
// 		count := 0
// 		for human := range humanChan {
// 			count++
// 			fmt.Printf("   📥 Received: %s (%s)\n", human.Name, human.ID)
// 			if count >= 3 { // Limit for demo
// 				break
// 			}
// 		}
// 		fmt.Printf("✅ Received %d humans via stream\n", count)
// 	}

// 	// Test UploadHumans (client streaming)
// 	fmt.Println("\n📤 Testing UploadHumans (client streaming)...")
// 	ctx4, cancel4 := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel4()

// 	// Create input channel for upload
// 	uploadChan := make(chan models.Human, 2)

// 	// Send some test humans synchronously first
// 	go func() {
// 		defer close(uploadChan)
// 		time.Sleep(50 * time.Millisecond) // Give client time to start
// 		uploadChan <- models.Human{
// 			ID:     "test-human-1",
// 			Name:   "Test Human 1",
// 			Height: 1.75,
// 		}
// 		time.Sleep(50 * time.Millisecond)
// 		uploadChan <- models.Human{
// 			ID:     "test-human-2",
// 			Name:   "Test Human 2",
// 			Height: 1.80,
// 		}
// 		time.Sleep(50 * time.Millisecond) // Give time before closing
// 	}()

// 	uploadResp, err := client.UploadHumans(ctx4, uploadChan)
// 	if err != nil {
// 		log.Printf("❌ UploadHumans failed: %v", err)
// 	} else {
// 		fmt.Printf("✅ Upload successful: %d humans uploaded\n", uploadResp.Count)
// 	}

// 	// Test ChatWithCharacters (bidirectional streaming)
// 	fmt.Println("\n💬 Testing ChatWithCharacters (bidirectional streaming)...")
// 	ctx5, cancel5 := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel5()

// 	// Create input channel for chat
// 	chatInputChan := make(chan models.ChatMessage, 2)

// 	// Start chat
// 	chatOutputChan, err := client.ChatWithCharacters(ctx5, chatInputChan)
// 	if err != nil {
// 		log.Printf("❌ ChatWithCharacters failed: %v", err)
// 	} else {
// 		// Send a few messages
// 		go func() {
// 			defer close(chatInputChan)
// 			time.Sleep(100 * time.Millisecond)
// 			chatInputChan <- models.ChatMessage{
// 				ID:        "msg-1",
// 				From:      "client",
// 				To:        "luke-skywalker",
// 				Message:   "Hello, Luke!",
// 				Timestamp: time.Now().Unix(),
// 			}
// 			time.Sleep(100 * time.Millisecond)
// 			chatInputChan <- models.ChatMessage{
// 				ID:        "msg-2",
// 				From:      "client",
// 				To:        "c-3po",
// 				Message:   "How are you, C-3PO?",
// 				Timestamp: time.Now().Unix(),
// 			}
// 		}()

// 		// Read responses
// 		responseCount := 0
// 		timeout := time.After(3 * time.Second)
// 		for {
// 			select {
// 			case response, ok := <-chatOutputChan:
// 				if !ok {
// 					goto chatDone
// 				}
// 				responseCount++
// 				fmt.Printf("   💬 Chat response: %s -> %s: %s\n", response.From, response.To, response.Message)
// 				if responseCount >= 2 {
// 					goto chatDone
// 				}
// 			case <-timeout:
// 				goto chatDone
// 			}
// 		}
// 	chatDone:
// 		fmt.Printf("✅ Chat completed: %d messages exchanged\n", responseCount)
// 	}

// 	fmt.Println("\n🎉 Client test completed successfully!")
// 	fmt.Println("=" + string(make([]rune, 50)))
// }
