// Package adapter contains auto-generated type adapters
// Generated from protobuf annotations - DO NOT EDIT
package adapter

import (
	"github.com/coso/models"
	pb "github.com/coso/generated/proto/v1"
)
// CreateUserResponseToProto converts models.CreateUserResponse to protobuf *pb.CreateUserResponse
func CreateUserResponseToProto(orig models.CreateUserResponse) *pb.CreateUserResponse {
	if IsZeroCreateUserResponse(orig) {
		return nil
	}

	proto := &pb.CreateUserResponse{
		Success: orig.Success,
		Message: orig.Message,
	}

	return proto
}

// CreateUserResponseFromProto converts protobuf *pb.CreateUserResponse to models.CreateUserResponse
func CreateUserResponseFromProto(proto *pb.CreateUserResponse) models.CreateUserResponse {
	if proto == nil {
		return models.CreateUserResponse{}
	}

	orig := models.CreateUserResponse{
		Success: proto.Success,
		Message: proto.Message,
	}

	return orig
}

// IsZeroCreateUserResponse checks if the struct is its zero value
func IsZeroCreateUserResponse(v models.CreateUserResponse) bool {
	// Check if all fields are zero values
	if v.Success {
		return false
	}
	if v.Message != "" {
		return false
	}
	return true
}

// CreateUserResponseSliceToProto converts []models.CreateUserResponse to []*pb.CreateUserResponse
func CreateUserResponseSliceToProto(orig []models.CreateUserResponse) []*pb.CreateUserResponse {
	if len(orig) == 0 {
		return nil
	}
	
	result := make([]*pb.CreateUserResponse, len(orig))
	for i, v := range orig {
		result[i] = CreateUserResponseToProto(v)
	}
	return result
}

// CreateUserResponseSliceFromProto converts []*pb.CreateUserResponse to []models.CreateUserResponse
func CreateUserResponseSliceFromProto(proto []*pb.CreateUserResponse) []models.CreateUserResponse {
	if len(proto) == 0 {
		return nil
	}
	
	result := make([]models.CreateUserResponse, len(proto))
	for i, v := range proto {
		result[i] = CreateUserResponseFromProto(v)
	}
	return result
}
// UserToProto converts models.User to protobuf *pb.User
func UserToProto(orig models.User) *pb.User {
	if IsZeroUser(orig) {
		return nil
	}

	proto := &pb.User{
		Id: orig.ID,
		Username: orig.Username,
		Email: orig.Email,
	}

	return proto
}

// UserFromProto converts protobuf *pb.User to models.User
func UserFromProto(proto *pb.User) models.User {
	if proto == nil {
		return models.User{}
	}

	orig := models.User{
		ID: proto.Id,
		Username: proto.Username,
		Email: proto.Email,
	}

	return orig
}

// IsZeroUser checks if the struct is its zero value
func IsZeroUser(v models.User) bool {
	// Check if all fields are zero values
	if v.ID != 0 {
		return false
	}
	if v.Username != "" {
		return false
	}
	if v.Email != "" {
		return false
	}
	return true
}

// UserSliceToProto converts []models.User to []*pb.User
func UserSliceToProto(orig []models.User) []*pb.User {
	if len(orig) == 0 {
		return nil
	}
	
	result := make([]*pb.User, len(orig))
	for i, v := range orig {
		result[i] = UserToProto(v)
	}
	return result
}

// UserSliceFromProto converts []*pb.User to []models.User
func UserSliceFromProto(proto []*pb.User) []models.User {
	if len(proto) == 0 {
		return nil
	}
	
	result := make([]models.User, len(proto))
	for i, v := range proto {
		result[i] = UserFromProto(v)
	}
	return result
}
// GetUserRequestToProto converts models.GetUserRequest to protobuf *pb.GetUserRequest
func GetUserRequestToProto(orig models.GetUserRequest) *pb.GetUserRequest {
	if IsZeroGetUserRequest(orig) {
		return nil
	}

	proto := &pb.GetUserRequest{
		Id: orig.ID,
	}

	return proto
}

// GetUserRequestFromProto converts protobuf *pb.GetUserRequest to models.GetUserRequest
func GetUserRequestFromProto(proto *pb.GetUserRequest) models.GetUserRequest {
	if proto == nil {
		return models.GetUserRequest{}
	}

	orig := models.GetUserRequest{
		ID: proto.Id,
	}

	return orig
}

// IsZeroGetUserRequest checks if the struct is its zero value
func IsZeroGetUserRequest(v models.GetUserRequest) bool {
	// Check if all fields are zero values
	if v.ID != 0 {
		return false
	}
	return true
}

// GetUserRequestSliceToProto converts []models.GetUserRequest to []*pb.GetUserRequest
func GetUserRequestSliceToProto(orig []models.GetUserRequest) []*pb.GetUserRequest {
	if len(orig) == 0 {
		return nil
	}
	
	result := make([]*pb.GetUserRequest, len(orig))
	for i, v := range orig {
		result[i] = GetUserRequestToProto(v)
	}
	return result
}

// GetUserRequestSliceFromProto converts []*pb.GetUserRequest to []models.GetUserRequest
func GetUserRequestSliceFromProto(proto []*pb.GetUserRequest) []models.GetUserRequest {
	if len(proto) == 0 {
		return nil
	}
	
	result := make([]models.GetUserRequest, len(proto))
	for i, v := range proto {
		result[i] = GetUserRequestFromProto(v)
	}
	return result
}
// CreateUserRequestToProto converts models.CreateUserRequest to protobuf *pb.CreateUserRequest
func CreateUserRequestToProto(orig models.CreateUserRequest) *pb.CreateUserRequest {
	if IsZeroCreateUserRequest(orig) {
		return nil
	}

	proto := &pb.CreateUserRequest{
		User: ConvertPointerToProto_User(orig.User),
	}

	return proto
}

// CreateUserRequestFromProto converts protobuf *pb.CreateUserRequest to models.CreateUserRequest
func CreateUserRequestFromProto(proto *pb.CreateUserRequest) models.CreateUserRequest {
	if proto == nil {
		return models.CreateUserRequest{}
	}

	orig := models.CreateUserRequest{
		User: ConvertPointerFromProto_User(proto.User),
	}

	return orig
}

// IsZeroCreateUserRequest checks if the struct is its zero value
func IsZeroCreateUserRequest(v models.CreateUserRequest) bool {
	// Check if all fields are zero values
	// For complex types, use reflection or custom logic
	if v.User != (*User{}) {
		return false
	}
	return true
}

// CreateUserRequestSliceToProto converts []models.CreateUserRequest to []*pb.CreateUserRequest
func CreateUserRequestSliceToProto(orig []models.CreateUserRequest) []*pb.CreateUserRequest {
	if len(orig) == 0 {
		return nil
	}
	
	result := make([]*pb.CreateUserRequest, len(orig))
	for i, v := range orig {
		result[i] = CreateUserRequestToProto(v)
	}
	return result
}

// CreateUserRequestSliceFromProto converts []*pb.CreateUserRequest to []models.CreateUserRequest
func CreateUserRequestSliceFromProto(proto []*pb.CreateUserRequest) []models.CreateUserRequest {
	if len(proto) == 0 {
		return nil
	}
	
	result := make([]models.CreateUserRequest, len(proto))
	for i, v := range proto {
		result[i] = CreateUserRequestFromProto(v)
	}
	return result
}
// ConvertPointerToProto_CreateUserResponse converts *models.CreateUserResponse to *pb.CreateUserResponse
func ConvertPointerToProto_CreateUserResponse(orig *models.CreateUserResponse) *pb.CreateUserResponse {
	if orig == nil {
		return nil
	}
	return CreateUserResponseToProto(*orig)
}

// ConvertPointerFromProto_CreateUserResponse converts *pb.CreateUserResponse to *models.CreateUserResponse
func ConvertPointerFromProto_CreateUserResponse(proto *pb.CreateUserResponse) *models.CreateUserResponse {
	if proto == nil {
		return nil
	}
	result := CreateUserResponseFromProto(proto)
	return &result
}
// ConvertPointerToProto_User converts *models.User to *pb.User
func ConvertPointerToProto_User(orig *models.User) *pb.User {
	if orig == nil {
		return nil
	}
	return UserToProto(*orig)
}

// ConvertPointerFromProto_User converts *pb.User to *models.User
func ConvertPointerFromProto_User(proto *pb.User) *models.User {
	if proto == nil {
		return nil
	}
	result := UserFromProto(proto)
	return &result
}
// ConvertPointerToProto_GetUserRequest converts *models.GetUserRequest to *pb.GetUserRequest
func ConvertPointerToProto_GetUserRequest(orig *models.GetUserRequest) *pb.GetUserRequest {
	if orig == nil {
		return nil
	}
	return GetUserRequestToProto(*orig)
}

// ConvertPointerFromProto_GetUserRequest converts *pb.GetUserRequest to *models.GetUserRequest
func ConvertPointerFromProto_GetUserRequest(proto *pb.GetUserRequest) *models.GetUserRequest {
	if proto == nil {
		return nil
	}
	result := GetUserRequestFromProto(proto)
	return &result
}
// ConvertPointerToProto_CreateUserRequest converts *models.CreateUserRequest to *pb.CreateUserRequest
func ConvertPointerToProto_CreateUserRequest(orig *models.CreateUserRequest) *pb.CreateUserRequest {
	if orig == nil {
		return nil
	}
	return CreateUserRequestToProto(*orig)
}

// ConvertPointerFromProto_CreateUserRequest converts *pb.CreateUserRequest to *models.CreateUserRequest
func ConvertPointerFromProto_CreateUserRequest(proto *pb.CreateUserRequest) *models.CreateUserRequest {
	if proto == nil {
		return nil
	}
	result := CreateUserRequestFromProto(proto)
	return &result
}

// Helper functions for pointer conversion
func stringPtrToValue(ptr *string) string { if ptr == nil { return "" }; return *ptr }
func valueToPtrString(val string) *string { if val == "" { return nil }; return &val }
func int32PtrToValue(ptr *int32) int32 { if ptr == nil { return 0 }; return *ptr }
func valueToPtrInt32(val int32) *int32 { if val == 0 { return nil }; return &val }
func int64PtrToValue(ptr *int64) int64 { if ptr == nil { return 0 }; return *ptr }
func valueToPtrInt64(val int64) *int64 { if val == 0 { return nil }; return &val }
func float32PtrToValue(ptr *float32) float32 { if ptr == nil { return 0 }; return *ptr }
func valueToPtrFloat32(val float32) *float32 { if val == 0 { return nil }; return &val }
func float64PtrToValue(ptr *float64) float64 { if ptr == nil { return 0 }; return *ptr }
func valueToPtrFloat64(val float64) *float64 { if val == 0 { return nil }; return &val }
func boolPtrToValue(ptr *bool) bool { if ptr == nil { return false }; return *ptr }
func valueToPtrBool(val bool) *bool { return &val }