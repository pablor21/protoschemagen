# protoschemagen

**Generate Protocol Buffers schemas and gRPC services from Go code using annotations.**

[![Go Reference](https://pkg.go.dev/badge/github.com/pablor21/protoschemagen.svg)](https://pkg.go.dev/github.com/pablor21/protoschemagen)
[![Go Report Card](https://goreportcard.com/badge/github.com/pablor21/protoschemagen)](https://goreportcard.com/report/github.com/pablor21/protoschemagen)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

protoschemagen is a powerful code generation tool that automatically creates Protocol Buffers schemas, gRPC services, and type-safe adapters from your existing Go code. Simply add annotations to your structs and interfaces, and let protoschemagen handle the rest!

## ✨ What Makes It Special?

- 🎯 **Annotation-driven** - Use simple Go comments to define protobuf behavior
- 🔄 **Type-safe adapters** - Automatic conversion between Go and protobuf types  
- 🚀 **Complete gRPC integration** - Server adapters, clients, and registration helpers
- 🌊 **Full streaming support** - Server, client, and bidirectional streaming
- 📁 **Flexible output** - Single file or multiple file generation strategies
- ⚡ **Zero boilerplate** - Focus on business logic, not protobuf plumbing

## 🚀 Quick Example

Transform this Go code:

```go
// @protobuf.message
type User struct {
    // @protobuf.field(number=1)
    ID string `json:"id"`
    
    // @protobuf.field(number=2)
    Name string `json:"name"`
    
    // @protobuf.field(number=3)
    Email string `json:"email"`
}

// @protobuf.service
type UserService interface {
    // @protobuf.rpc
    GetUser(ctx context.Context, req GetUserRequest) (*User, error)
    
    // @protobuf.rpc
    CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
}
```

Into this protobuf schema + complete gRPC implementation:

```protobuf
syntax = "proto3";
package user.v1;

message User {
  string id = 1;
  string name = 2;
  string email = 3;
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc CreateUser(CreateUserRequest) returns (User);
}
```

**Plus:** Type conversion functions, gRPC adapters, client wrappers, and registration helpers!

## 📖 Documentation

Our comprehensive wiki covers everything you need to know:

### 🎯 **[Getting Started](wiki/Getting-Started.md)**
Installation, first schema generation, and basic concepts

### 📝 **[Annotations Guide](wiki/Annotations-Guide.md)**  
Complete reference for all protobuf annotations and their usage

### ⚙️ **[Configuration](wiki/Configuration.md)**
Detailed configuration options and customization

### 🔧 **[Generated Code](wiki/Generated-Code.md)**
Understanding what protoschemagen creates and how to use it

### 🚀 **[Examples](wiki/Examples.md)**
Real-world examples from simple CRUD to complex streaming services

### 🎛️ **[Advanced Features](wiki/Advanced-Features.md)**
Streaming, custom templates, plugins, and integration patterns

### 🛠️ **[Troubleshooting](wiki/Troubleshooting.md)**
Common issues, debugging techniques, and solutions

### 📚 **[API Reference](wiki/API-Reference.md)**
Complete reference for annotations, configuration, and CLI

## ⚡ Quick Start

### 1. Install protoschemagen

```bash
go install github.com/pablor21/protoschemagen@latest
```

### 2. Install dependencies

```bash
# Install protoc and Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 3. Add annotations to your Go code

```go
// models/user.go
package models

// @protobuf.message
type User struct {
    // @protobuf.field(number=1)
    ID string `json:"id"`
    
    // @protobuf.field(number=2) 
    Name string `json:"name"`
}

// @protobuf.service
type UserService interface {
    // @protobuf.rpc
    GetUser(ctx context.Context, id string) (*User, error)
}
```

### 4. Create configuration

```yaml
# protoschemagen.yml
packages:
  - "./models/**"
  
generate:
  - protobuf

plugins:
  protobuf:
    enabled: true
    output: "schema/user.proto"
    package: "user.v1"
    options:
      go_package: "github.com/example/user/proto/v1"
    
    generate_stubs:
      enabled: true
      adapter_package: "generated/adapter"
```

### 5. Generate everything

```bash
protoschemagen -config=protoschemagen.yml
```

### 6. Use in your server

```go
package main

import (
    "github.com/example/user/models"
    "github.com/example/user/generated/adapter"
)

func main() {
    // Your business logic
    userSvc := &MyUserService{}
    
    // Create gRPC server
    server := grpc.NewServer()
    
    // Register with generated adapter (one line!)
    adapter.RegisterUserService(server, userSvc)
    
    // Start server
    server.Serve(listener)
}
```

That's it! You now have a complete gRPC service with type-safe conversion between your Go types and protobuf.

## 🌟 Key Features

### 🎯 **Annotation-Driven Design**
- Simple Go comments control protobuf generation
- No separate `.proto` files to maintain
- Keep your Go code as the single source of truth

### 🔄 **Smart Type Conversion**
- Automatic conversion between Go and protobuf types
- Handles complex nested structs, slices, maps, and pointers
- Support for `time.Time`, `time.Duration`, and custom types

### 🚀 **Complete gRPC Integration** 
- Generated service adapters implement gRPC interfaces
- Client wrappers provide Go-native interfaces
- Registration helpers for easy server setup

### 🌊 **Advanced Streaming Support**
- Server streaming: `chan ResponseType`
- Client streaming: `chan RequestType`  
- Bidirectional streaming: `chan RequestType, chan ResponseType`
- All with automatic adapter generation

### 📁 **Flexible Generation Strategies**
- **Single file:** All types in one `.proto` file
- **Follow:** Separate files following Go package structure
- **Category:** Group by type (messages, services, enums)

### ⚡ **Developer Experience**
- Auto field numbering (optional)
- Incremental builds and caching
- Rich error messages with line numbers
- IDE integration support

## 🎮 Live Examples

Check out our working examples:

```bash
# Clone the repository
git clone https://github.com/pablor21/protoschemagen
cd protoschemagen

# Simple CRUD service
cd examples/simple
go generate && go run server/main.go

# Complex streaming service (Star Wars API)
cd examples/starwars  
go generate && go run server/main.go

# Multi-service pet store
cd examples/petstore
go generate && go run server/main.go
```

Each example includes:
- ✅ Complete Go service implementations
- ✅ Generated protobuf schemas
- ✅ Working gRPC servers with adapters
- ✅ Client examples
- ✅ Detailed documentation

## 🛠️ Why protoschemagen?

### Traditional Approach 😓
```
1. Write Go structs
2. Manually create .proto files  
3. Keep both in sync manually
4. Write conversion code by hand
5. Handle type mismatches
6. Repeat for every change
```

### With protoschemagen 🎉
```
1. Write Go structs with annotations
2. Run protoschemagen
3. Everything is generated and in sync!
```

### The Result
- **90% less boilerplate** - Focus on business logic, not plumbing
- **Zero drift** - Go code and protobuf always in sync
- **Type safety** - Compile-time guarantees for conversions
- **Maintainability** - Single source of truth in Go code

## 🤝 Contributing

We welcome contributions! Whether you want to:
- 🐛 Report bugs
- 💡 Suggest features  
- 📖 Improve documentation
- 🔧 Submit code changes

Check our [contribution guidelines](CONTRIBUTING.md) and join our community!

## 📄 License

protoschemagen is released under the [MIT License](LICENSE).

## 🙏 Acknowledgments

Special thanks to:
- The [Protocol Buffers](https://developers.google.com/protocol-buffers) team
- The [gRPC](https://grpc.io/) community  
<!-- - All our [contributors](https://github.com/pablor21/protoschemagen/contributors) -->

---

## 🎯 Ready to Get Started?

**👉 [Check out the Getting Started guide](wiki/Getting-Started.md) and transform your Go code into powerful gRPC services in minutes!**

Have questions? Check our [FAQ](wiki/Troubleshooting.md#common-issues) or [open a discussion](https://github.com/pablor21/protoschemagen/discussions).

*Built with ❤️ for the Go community*
