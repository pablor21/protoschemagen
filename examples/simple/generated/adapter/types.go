// Package adapter contains auto-generated type adapters
// Generated from protobuf annotations - DO NOT EDIT
package adapter

import (
	"github.com/coso/models"
	pb "github.com/coso/generated/proto/v1"
)
// UserToProto converts models.User to protobuf *pb.User
func UserToProto(orig models.User) *pb.User {
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