# Star Wars gRPC Service Example

This example demonstrates a complete protobuf-based gRPC service with type-safe adapters generated from Go struct annotations.

## What's Included

- **In-Memory Star Wars Database**: Sample humans (Luke, Leia, Han) and droids (C-3PO, R2-D2)
- **Generated gRPC Service**: Protobuf schema and Go code generated from annotations
- **Type-Safe Adapters**: Automatic conversion between Go models and protobuf types
- **Generic Type Support**: Handles complex types like `Response[Human]` correctly
- **Map/Slice Conversions**: Proper handling of complex nested data structures

## Files

### Source Code with Annotations
- `models/` - Go structs with protobuf annotations
- `models/service.go` - Service interface with `@proto.service` and `@proto.rpc` annotations

### Configuration
- `test-config.yml` - Configuration for protobuf schema generation

### Generated Artifacts
- `schema/schema.proto` - Generated protobuf schema
- `generated/` - Generated Go stubs from protobuf

### Examples
- `example_server.go` - Shows how to implement the generated service interface
- `example_client.go` - Shows how to use the generated client interface

## Step 1: Generate Protobuf Schema

Generate the protobuf schema from Go annotations:

```bash
# Build the protoschemagen tool
cd ../..
go build -o protoschemagen .

# Generate schema from Go annotations
cd examples/starwars
../../protoschemagen -config test-config.yml .
```

This reads the Go structs and interfaces with `@proto.*` annotations and generates `schema/schema.proto`.

## Step 2: Generate Go Stubs

Generate Go gRPC stubs from the protobuf schema using the standard `protoc` approach:

```bash
# Run the stub generation script
./generate_stubs.sh
```

This executes the equivalent of:
```bash
cd schema
protoc -I=. --go_out=. --go-grpc_out=. schema.proto
```

And produces:
- `generated/schema.pb.go` - Generated message types
- `generated/schema_grpc.pb.go` - Generated service interfaces and client

## Step 3: Implement the Service

The generated `schema_grpc.pb.go` contains:

### Service Interface (to implement)
```go
type StarWarsServiceServer interface {
    GetHuman(context.Context, *GetHumanRequest) (*Human, error)
    GetDroid(context.Context, *GetDroidRequest) (*Droid, error)
    StreamHumans(*StreamHumansRequest, grpc.ServerStreamingServer[Human]) error
    UploadHumans(grpc.ClientStreamingServer[Human, UploadHumansResponse]) error
    ChatWithCharacters(grpc.BidiStreamingServer[ChatMessage, ChatMessage]) error
    mustEmbedUnimplementedStarWarsServiceServer()
}
```

### Client Interface (for calling the service)
```go
type StarWarsServiceClient interface {
    GetHuman(ctx context.Context, in *GetHumanRequest, opts ...grpc.CallOption) (*Human, error)
    GetDroid(ctx context.Context, in *GetDroidRequest, opts ...grpc.CallOption) (*Droid, error)
    StreamHumans(ctx context.Context, in *StreamHumansRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[Human], error)
    UploadHumans(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[Human, UploadHumansResponse], error)
    ChatWithCharacters(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[ChatMessage, ChatMessage], error)
}
```

## Step 4: Test the Implementation

Build and run the example:

```bash
# Terminal 1: Start the server
go run example_server.go

# Terminal 2: Run the client
go run example_client.go
```

## Streaming Support

The generated service supports all gRPC streaming patterns:

### Unary RPC
```protobuf
rpc GetHuman(GetHumanRequest) returns (Human);
```

### Server Streaming  
```protobuf
rpc StreamHumans(StreamHumansRequest) returns (stream Human);
```

### Client Streaming
```protobuf
rpc UploadHumans(stream Human) returns (UploadHumansResponse);
```

### Bidirectional Streaming
```protobuf
rpc ChatWithCharacters(stream ChatMessage) returns (stream ChatMessage);
```

## Key Benefits

1. **Type Safety**: Generated code is strongly typed
2. **Interface Driven**: Clear separation between interface and implementation  
3. **Standard Tools**: Uses standard `protoc` for stub generation
4. **Forward Compatible**: Embedding `UnimplementedServer` ensures compatibility
5. **Complete Coverage**: Supports all gRPC patterns (unary, streaming)

The generated stubs provide the foundation - you implement the business logic!