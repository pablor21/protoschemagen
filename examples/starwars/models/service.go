package models

import "context"

// User-facing StarWars service
// @namespace("api")
// @service name:"StarWarsService"
type StarWarsService interface {
	// Get a human by ID
	// @rpc name:"GetHuman" input:"GetHumanRequest" output:"Human"
	GetHuman(ctx context.Context, req GetHumanRequest) (Human, error)

	// Get a droid by ID
	// @rpc name:"GetDroid" input:"GetDroidRequest" output:"Droid"
	GetDroid(ctx context.Context, req GetDroidRequest) (Droid, error)

	// Stream humans - server streaming
	// @rpc name:"StreamHumans" input:"StreamHumansRequest" output:"Human" server_streaming:"true"
	StreamHumans(ctx context.Context, req StreamHumansRequest) ([]Human, error)

	// Upload human data - client streaming
	// @rpc name:"UploadHumans" input:"Human" output:"UploadHumansResponse" client_streaming:"true"
	UploadHumans(ctx context.Context, humans []Human) (UploadHumansResponse, error)
	// Chat with characters - bidirectional streaming
	// @rpc name:"ChatWithCharacters" input:"ChatMessage" output:"ChatMessage" client_streaming:"true" server_streaming:"true"
	ChatWithCharacters(ctx context.Context, messages []ChatMessage) ([]ChatMessage, error)
}

// Request messages
// @namespace("api")
// @message name:"GetHumanRequest"
type GetHumanRequest struct {
	// @field number:1
	ID string
}

// @namespace("api")
// @message name:"GetDroidRequest"
type GetDroidRequest struct {
	// @field number:1
	ID string
}

// @namespace("api")
// @message name:"StreamHumansRequest"
type StreamHumansRequest struct {
	// @field number:1
	Limit int32
	// @field number:2
	Offset int32
}

// @namespace("api")
// @message name:"UploadHumansResponse"
type UploadHumansResponse struct {
	// @field number:1
	Count int32
	// @field number:2
	Status string
}

// @namespace("api")
// @message name:"ChatMessage"
type ChatMessage struct {
	// @field number:1
	ID string
	// @field number:2
	From string
	// @field number:3
	To string
	// @field number:4
	Message string
	// @field number:5
	Timestamp int64
}
