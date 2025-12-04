package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pablor21/gonnotation/parser"
	"github.com/pablor21/goschemagen"
	"github.com/pablor21/protoschemagen/plugin"
)

// UserProfile represents a user's profile information
// This struct contains personal and contact details
type UserProfile struct {
	// The unique identifier for the user
	// @proto.field(number=1)
	UserID string `json:"user_id"`

	// Full name of the user as it appears on official documents
	// @proto.field(number=2)
	FullName string `json:"full_name"`

	// Primary email address for communication
	// @proto.field(number=3)
	Email string `json:"email"`
}

// TestCommentExtraction verifies that comments are extracted and used as descriptions
func TestCommentExtraction(t *testing.T) {
	// Create a temporary directory and test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// Create a simple Go source with comments
	source := `package testpkg

// UserProfile represents a user's profile information
// This struct contains personal and contact details
type UserProfile struct {
	// The unique identifier for the user
	UserID string
	
	// Full name of the user as it appears on official documents  
	FullName string
	
	// Primary email address for communication
	Email string
}`

	if err := os.WriteFile(testFile, []byte(source), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	p := parser.NewParser()
	err := p.ParsePackages([]string{tmpDir})
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	structs := p.ExtractStructs()
	if len(structs) == 0 {
		t.Fatal("No structs found")
	}

	userProfile := structs[0]
	if userProfile.Name != "UserProfile" {
		t.Fatalf("Expected struct name 'UserProfile', got '%s'", userProfile.Name)
	}

	// Test struct comment extraction
	if userProfile.Comment == "" {
		t.Error("Struct comment should not be empty")
	}

	if !strings.Contains(userProfile.Comment, "UserProfile represents") {
		t.Errorf("Expected struct comment to contain description, got: '%s'", userProfile.Comment)
	}

	// Test field comment extraction
	userIDField := findField(userProfile.Fields, "UserID")
	if userIDField == nil {
		t.Fatal("UserID field not found")
	}

	if userIDField.Comment == "" {
		t.Error("UserID field comment should not be empty")
	}

	if !strings.Contains(userIDField.Comment, "unique identifier") {
		t.Errorf("Expected UserID comment to contain description, got: '%s'", userIDField.Comment)
	}

	// Test goschemagen description helper with UseCommentsAsDescription enabled
	config := &goschemagen.Config{
		UseCommentsAsDescription: &[]bool{true}[0], // Pointer to true
	}

	structDesc := config.GetStructDescription(userProfile)
	if structDesc == "" {
		t.Error("Should get struct description when UseCommentsAsDescription is true")
	}

	fieldDesc := config.GetFieldDescription(userIDField)
	if fieldDesc == "" {
		t.Error("Should get field description when UseCommentsAsDescription is true")
	}

	// Test with UseCommentsAsDescription disabled
	config.UseCommentsAsDescription = &[]bool{false}[0] // Pointer to false

	structDescDisabled := config.GetStructDescription(userProfile)
	if structDescDisabled != "" {
		t.Error("Should not get struct description when UseCommentsAsDescription is false")
	}

	fieldDescDisabled := config.GetFieldDescription(userIDField)
	if fieldDescDisabled != "" {
		t.Error("Should not get field description when UseCommentsAsDescription is false")
	}
}

// TestProtobufGenerationWithComments verifies protobuf generation uses comments when enabled
func TestProtobufGenerationWithComments(t *testing.T) {
	// Create a temporary directory and test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	source := `package testpkg

// UserProfile represents a user's profile information  
type UserProfile struct {
	// The unique identifier for the user
	// @proto.field(number=1)
	UserID string
	
	// Primary email address for communication
	// @proto.field(number=2) 
	Email string
}`

	if err := os.WriteFile(testFile, []byte(source), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	p := parser.NewParser()
	err := p.ParsePackages([]string{tmpDir})
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	structs := p.ExtractStructs()
	enums := p.ExtractEnums()
	interfaces := p.ExtractInterfaces()
	functions := p.ExtractFunctions()

	// Create generation context
	coreConfig := &parser.CoreConfig{}
	ctx := parser.NewGenerationContextWithInterfaces(
		p, structs, interfaces, enums, functions,
		structs, interfaces, enums, functions,
		coreConfig, nil, nil, make(map[string]bool),
	)
	if ctx == nil {
		t.Fatal("Failed to build generation context")
	} // Test with comments enabled
	config := &plugin.Config{
		Package: "test.v1",
		Syntax:  "proto3",
	}
	config.UseCommentsAsDescription = &[]bool{true}[0] // Enable comments

	protobufPlugin := plugin.NewPlugin(config)
	schema, err := protobufPlugin.Generate(ctx)
	if err != nil {
		t.Fatalf("Failed to generate protobuf schema: %v", err)
	}

	schemaStr := string(schema)
	t.Logf("Generated schema:\n%s", schemaStr)

	// Verify comments are included in generated proto
	if !strings.Contains(schemaStr, "UserProfile represents") {
		t.Error("Generated schema should include struct comment")
	}

	if !strings.Contains(schemaStr, "unique identifier") {
		t.Error("Generated schema should include field comments")
	}

	// Test with comments disabled
	config.UseCommentsAsDescription = &[]bool{false}[0] // Disable comments

	protobufPluginNoComments := plugin.NewPlugin(config)
	schemaNoComments, err := protobufPluginNoComments.Generate(ctx)
	if err != nil {
		t.Fatalf("Failed to generate protobuf schema without comments: %v", err)
	}

	schemaNoCommentsStr := string(schemaNoComments)
	t.Logf("Generated schema without comments:\n%s", schemaNoCommentsStr)

	// Comments should not be present when disabled
	if strings.Contains(schemaNoCommentsStr, "UserProfile represents") {
		t.Error("Generated schema should not include struct comment when disabled")
	}
}

// Helper function to find a field by name
func findField(fields []*parser.FieldInfo, name string) *parser.FieldInfo {
	for _, field := range fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}
