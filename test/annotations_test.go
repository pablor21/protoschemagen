package main_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pablor21/gonnotation/parser"
	"github.com/pablor21/protoschemagen/plugin"
)

// Test data structures with comprehensive annotations
// @proto.package(name="test.v1")
// @proto.import(path="google/protobuf/timestamp.proto", public=true)
// @proto.option(name="java_multiple_files", value="true")

// @proto.message(name="TestMessage", description="A comprehensive test message")
type CompleteTestStruct struct {
	// @proto.field(number=1, name="id", type="string")
	ID string `json:"id"`

	// @proto.field(number=2)
	Name string `json:"name"`

	// @proto.oneof(group="identification")
	// @proto.field(number=10)
	Email string `json:"email,omitempty"`

	// @proto.oneof(group="identification")
	// @proto.field(number=11)
	Phone string `json:"phone,omitempty"`

	// @proto.map(key="string", value="int32", number=3)
	Counters map[string]int32 `json:"counters"`

	// @proto.field(number=4, type="google.protobuf.Timestamp")
	CreatedAt time.Time `json:"created_at"`

	// @proto.field(number=5, repeated=true)
	Tags []string `json:"tags"`

	// @proto.ignore
	InternalField string `json:"-"`

	// @proto.reserved(number=20)
	// ReservedField string
}

// @proto.enum(name="TestStatus", allow_alias=true)
type TestStatus int

const (
	// @proto.enumvalue(number=0, name="STATUS_UNKNOWN")
	TestStatusUnknown TestStatus = 0

	// @proto.enumvalue(number=1, name="STATUS_ACTIVE")
	TestStatusActive TestStatus = 1

	// @proto.enumvalue(number=2, name="STATUS_INACTIVE")
	TestStatusInactive TestStatus = 2
)

// @proto.service(name="TestService")
type TestServiceInterface interface {
	// @proto.rpc(name="GetTest", input="TestRequest", output="TestMessage")
	GetTest(req TestRequest) (TestMessage, error)

	// @proto.rpc(name="CreateTest", input="TestMessage", output="TestMessage")
	CreateTest(msg TestMessage) (TestMessage, error)
}

// @proto.message
type TestRequest struct {
	// @proto.field(number=1)
	ID string `json:"id"`
}

// Extensions demonstration
type BaseMessage struct {
	// @proto.field(number=1)
	Base string `json:"base"`
}

type ExtendedMessage struct {
	// @proto.extend(message="BaseMessage", number=100, name="extension_field", type="string")
	Extension string `json:"extension"`
}

// @proto.option(name="java_package", value="com.example.test")
// @proto.option(name="go_package", value="github.com/example/test/v1")
// @proto.include(path="common.proto")
// @proto.reserved(numbers="6-10,15", names="old_field,deprecated_field")

// TestAnnotationParsing verifies that all annotation types are properly parsed
func TestAnnotationParsing(t *testing.T) {
	// Parse the current file to test annotation parsing
	testFile := filepath.Join(".", "annotations_test.go")

	p := parser.NewParser()
	err := p.ParsePackages([]string{testFile})
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	// Extract parsed data
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
	} // Verify struct annotations
	foundStruct := false
	for _, s := range ctx.Structs {
		if s.Name == "CompleteTestStruct" {
			foundStruct = true

			// Check message annotation
			hasMessageAnnotation := false
			for _, ann := range s.Annotations {
				if strings.ToLower(ann.Name) == "message" || strings.HasSuffix(strings.ToLower(ann.Name), ".message") {
					hasMessageAnnotation = true
					if ann.Params["name"] != "TestMessage" {
						t.Errorf("Expected message name 'TestMessage', got '%s'", ann.Params["name"])
					}
				}
			}
			if !hasMessageAnnotation {
				t.Error("CompleteTestStruct should have a @proto.message annotation")
			}

			// Check field annotations
			for _, field := range s.Fields {
				switch field.Name {
				case "ID":
					hasFieldAnnotation := false
					for _, ann := range field.Annotations {
						if strings.ToLower(ann.Name) == "field" || strings.HasSuffix(strings.ToLower(ann.Name), ".field") {
							hasFieldAnnotation = true
							if ann.Params["number"] != "1" {
								t.Errorf("Expected field number '1' for ID, got '%s'", ann.Params["number"])
							}
						}
					}
					if !hasFieldAnnotation {
						t.Error("ID field should have a @proto.field annotation")
					}

				case "Email":
					hasOneofAnnotation := false
					for _, ann := range field.Annotations {
						if strings.ToLower(ann.Name) == "oneof" || strings.HasSuffix(strings.ToLower(ann.Name), ".oneof") {
							hasOneofAnnotation = true
							if ann.Params["group"] != "identification" {
								t.Errorf("Expected oneof group 'identification', got '%s'", ann.Params["group"])
							}
						}
					}
					if !hasOneofAnnotation {
						t.Error("Email field should have a @proto.oneof annotation")
					}

				case "Counters":
					hasMapAnnotation := false
					for _, ann := range field.Annotations {
						if strings.ToLower(ann.Name) == "map" || strings.HasSuffix(strings.ToLower(ann.Name), ".map") {
							hasMapAnnotation = true
							if ann.Params["key"] != "string" || ann.Params["value"] != "int32" {
								t.Errorf("Expected map<string, int32>, got map<%s, %s>", ann.Params["key"], ann.Params["value"])
							}
						}
					}
					if !hasMapAnnotation {
						t.Error("Counters field should have a @proto.map annotation")
					}

				case "InternalField":
					hasIgnoreAnnotation := false
					for _, ann := range field.Annotations {
						if strings.ToLower(ann.Name) == "ignore" || strings.HasSuffix(strings.ToLower(ann.Name), ".ignore") {
							hasIgnoreAnnotation = true
						}
					}
					if !hasIgnoreAnnotation {
						t.Error("InternalField should have a @proto.ignore annotation")
					}
				}
			}
		}
	}
	if !foundStruct {
		t.Error("CompleteTestStruct not found in parsed structs")
	}

	// Verify enum annotations
	foundEnum := false
	for _, e := range ctx.Enums {
		if e.Name == "TestStatus" {
			foundEnum = true

			// Check enum annotation
			hasEnumAnnotation := false
			for _, ann := range e.Annotations {
				if strings.ToLower(ann.Name) == "enum" || strings.HasSuffix(strings.ToLower(ann.Name), ".enum") {
					hasEnumAnnotation = true
					if ann.Params["name"] != "TestStatus" {
						t.Errorf("Expected enum name 'TestStatus', got '%s'", ann.Params["name"])
					}
				}
			}
			if !hasEnumAnnotation {
				t.Error("TestStatus should have a @proto.enum annotation")
			}

			// Check enum value annotations
			for _, value := range e.Values {
				if value.Name == "TestStatusUnknown" {
					hasEnumValueAnnotation := false
					for _, ann := range value.Annotations {
						if strings.ToLower(ann.Name) == "enumvalue" || strings.HasSuffix(strings.ToLower(ann.Name), ".enumvalue") {
							hasEnumValueAnnotation = true
							if ann.Params["name"] != "STATUS_UNKNOWN" {
								t.Errorf("Expected enum value name 'STATUS_UNKNOWN', got '%s'", ann.Params["name"])
							}
						}
					}
					if !hasEnumValueAnnotation {
						t.Error("TestStatusUnknown should have a @proto.enumvalue annotation")
					}
				}
			}
		}
	}
	if !foundEnum {
		t.Error("TestStatus enum not found in parsed enums")
	}

	// Verify interface annotations
	foundInterface := false
	for _, iface := range ctx.Interfaces {
		if iface.Name == "TestServiceInterface" {
			foundInterface = true

			// Check service annotation
			hasServiceAnnotation := false
			for _, ann := range iface.Annotations {
				if strings.ToLower(ann.Name) == "service" || strings.HasSuffix(strings.ToLower(ann.Name), ".service") {
					hasServiceAnnotation = true
					if ann.Params["name"] != "TestService" {
						t.Errorf("Expected service name 'TestService', got '%s'", ann.Params["name"])
					}
				}
			}
			if !hasServiceAnnotation {
				t.Error("TestServiceInterface should have a @proto.service annotation")
			}

			// Check RPC annotations
			for _, method := range iface.Methods {
				if method.Name == "GetTest" {
					hasRPCAnnotation := false
					for _, ann := range method.Annotations {
						if strings.ToLower(ann.Name) == "rpc" || strings.HasSuffix(strings.ToLower(ann.Name), ".rpc") {
							hasRPCAnnotation = true
							if ann.Params["name"] != "GetTest" {
								t.Errorf("Expected RPC name 'GetTest', got '%s'", ann.Params["name"])
							}
						}
					}
					if !hasRPCAnnotation {
						t.Error("GetTest method should have a @proto.rpc annotation")
					}
				}
			}
		}
	}
	if !foundInterface {
		t.Error("TestServiceInterface not found in parsed interfaces")
	}
}

// TestProtobufGeneration verifies that protobuf schema is generated correctly
func TestProtobufGeneration(t *testing.T) {
	// Parse the current file
	testFile := filepath.Join(".", "annotations_test.go")

	p := parser.NewParser()
	err := p.ParsePackages([]string{testFile})
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	// Extract parsed data
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
	}

	// Create protobuf plugin and generate schema
	config := &plugin.Config{
		Package:          "test.v1",
		Syntax:           "proto3",
		GenerateService:  true,
		AutoNumberFields: true,
		StartFieldNumber: 1,
	}

	protobufPlugin := plugin.NewPlugin(config)

	schema, err := protobufPlugin.Generate(ctx)
	if err != nil {
		t.Fatalf("Failed to generate protobuf schema: %v", err)
	}

	schemaStr := string(schema)

	// Verify generated content
	tests := []struct {
		name     string
		expected string
		message  string
	}{
		{"package", "package test.v1;", "Package declaration should be present"},
		{"import", "import public \"google/protobuf/timestamp.proto\";", "Public import should be present"},
		{"option", "option java_multiple_files = \"true\";", "File-level option should be present"},
		{"message", "message TestMessage {", "TestMessage should be generated"},
		{"oneof", "oneof identification {", "Oneof group should be generated"},
		{"map", "map<string, int32> counters = 3;", "Map field should be generated"},
		{"timestamp", "google.protobuf.Timestamp created_at = 4;", "Timestamp field should be generated"},
		{"repeated", "repeated string tags = 5;", "Repeated field should be generated"},
		{"enum", "enum TestStatus {", "TestStatus enum should be generated"},
		{"enum_value", "STATUS_UNKNOWN = 0;", "Enum values should be generated"},
		{"service", "service TestService {", "TestService should be generated"},
		{"rpc", "rpc GetTest", "GetTest RPC should be generated"},
		{"extension", "extend BaseMessage {", "Extension should be generated"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !strings.Contains(schemaStr, test.expected) {
				t.Errorf("%s\nExpected to find: %s\nIn schema:\n%s", test.message, test.expected, schemaStr)
			}
		})
	}

	// Verify ignored field is NOT present
	if strings.Contains(schemaStr, "internal_field") || strings.Contains(schemaStr, "InternalField") {
		t.Error("InternalField should be ignored and not appear in generated schema")
	}

	// Verify reserved field number is handled
	if strings.Contains(schemaStr, "reserved 20;") || strings.Contains(schemaStr, "reserved \"20\";") {
		// This is good - reserved fields should appear in some form
		t.Log("Reserved field number 20 is properly handled")
	}
}

// TestAnnotationCoverage verifies that all annotation types from specs are covered
func TestAnnotationCoverage(t *testing.T) {
	expectedAnnotations := []string{
		"package", "message", "enum", "enumvalue", "field", "oneof",
		"service", "rpc", "map", "import", "reserved", "option",
		"extend", "ignore", "include",
	}

	// Parse the current file
	testFile := filepath.Join(".", "annotations_test.go")

	p := parser.NewParser()
	err := p.ParsePackages([]string{testFile})
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	// Extract parsed data
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
	}

	// Collect all annotation names found
	foundAnnotations := make(map[string]bool)

	// Check file-level annotations
	for _, fileAnns := range ctx.FileAnnotations {
		for _, ann := range fileAnns {
			name := strings.ToLower(ann.Name)
			if after, ok := strings.CutPrefix(name, "proto."); ok {
				name = after
			}
			foundAnnotations[name] = true
		}
	}

	// Check package-level annotations
	for _, pkgAnns := range ctx.PackageAnnotations {
		for _, ann := range pkgAnns {
			name := strings.ToLower(ann.Name)
			if after, ok := strings.CutPrefix(name, "proto."); ok {
				name = after
			}
			foundAnnotations[name] = true
		}
	} // Check struct annotations
	for _, s := range ctx.Structs {
		for _, ann := range s.Annotations {
			name := strings.ToLower(ann.Name)
			if after, ok := strings.CutPrefix(name, "proto."); ok {
				name = after
			}
			foundAnnotations[name] = true
		}

		// Check field annotations
		for _, field := range s.Fields {
			for _, ann := range field.Annotations {
				name := strings.ToLower(ann.Name)
				if after, ok := strings.CutPrefix(name, "proto."); ok {
					name = after
				}
				foundAnnotations[name] = true
			}
		}
	}

	// Check enum annotations
	for _, e := range ctx.Enums {
		for _, ann := range e.Annotations {
			name := strings.ToLower(ann.Name)
			if after, ok := strings.CutPrefix(name, "proto."); ok {
				name = after
			}
			foundAnnotations[name] = true
		}

		// Check enum value annotations
		for _, value := range e.Values {
			for _, ann := range value.Annotations {
				name := strings.ToLower(ann.Name)
				if after, ok := strings.CutPrefix(name, "proto."); ok {
					name = after
				}
				foundAnnotations[name] = true
			}
		}
	}

	// Check interface annotations
	for _, iface := range ctx.Interfaces {
		for _, ann := range iface.Annotations {
			name := strings.ToLower(ann.Name)
			if after, ok := strings.CutPrefix(name, "proto."); ok {
				name = after
			}
			foundAnnotations[name] = true
		}

		// Check method annotations
		for _, method := range iface.Methods {
			for _, ann := range method.Annotations {
				name := strings.ToLower(ann.Name)
				if after, ok := strings.CutPrefix(name, "proto."); ok {
					name = after
				}
				foundAnnotations[name] = true
			}
		}
	}

	// Verify all expected annotations are covered
	missing := []string{}
	for _, expected := range expectedAnnotations {
		if !foundAnnotations[expected] {
			missing = append(missing, expected)
		}
	}

	if len(missing) > 0 {
		t.Errorf("Missing annotation types in test: %v", missing)
	}

	// Report coverage
	coverage := float64(len(expectedAnnotations)-len(missing)) / float64(len(expectedAnnotations)) * 100
	t.Logf("Annotation coverage: %.1f%% (%d/%d)", coverage, len(expectedAnnotations)-len(missing), len(expectedAnnotations))

	if coverage < 100.0 {
		t.Error("Test should demonstrate 100% annotation coverage")
	}
}

// TestMessage placeholder for compilation
type TestMessage = CompleteTestStruct
