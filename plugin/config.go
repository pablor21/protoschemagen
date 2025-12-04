package plugin

import (
	"embed"

	"github.com/pablor21/gonnotation/parser"
	"github.com/pablor21/goschemagen"
	"gopkg.in/yaml.v3"
)

//go:embed protobuf.generators.yml
var DefaultConfigFs embed.FS
var DefaultConfigFile = "protobuf.generators.yml"
var DefaultConfigData, _ = DefaultConfigFs.ReadFile(DefaultConfigFile)

// Config holds Protobuf-specific configuration
type Config struct {
	// Output settings
	// Output is the directory where files will be generated
	Output string `yaml:"output"`

	// Output file name pattern (supports {schema_name}, {package}, {namespace})
	// Example: "{schema_name}.proto" or "{package}.proto"
	OutputFileName string `yaml:"output_file_name"`

	// Output formats to generate (default: ["proto"])
	// Supported: proto, json-schema, markdown, typescript, descriptor
	OutputFormats []string `yaml:"output_formats"`

	// Generation strategy: "single", "follow", "package", "namespace"
	GenerationStrategy parser.GenStrategy `yaml:"generation_strategy"`

	// Deprecated: use generation_strategy instead
	GenStrategy parser.GenStrategy `yaml:"strategy,omitempty"`

	// Protobuf-specific features
	Syntax          string `yaml:"syntax"`           // proto2 or proto3 (default: proto3)
	Package         string `yaml:"package"`          // Protobuf package name
	GoPackage       string `yaml:"go_package"`       // Go package import path
	JavaPackage     string `yaml:"java_package"`     // Java package name
	JavaOuterClass  string `yaml:"java_outer_class"` // Java outer class name
	OptimizeFor     string `yaml:"optimize_for"`     // SPEED, CODE_SIZE, LITE_RUNTIME
	GenerateService bool   `yaml:"generate_service"` // Generate gRPC service definitions

	// Field numbering
	AutoNumberFields bool     `yaml:"auto_number_fields"` // Auto-assign field numbers
	StartFieldNumber int      `yaml:"start_field_number"` // Starting field number (default: 1)
	ReservedNumbers  []int    `yaml:"reserved_numbers"`   // Reserved field numbers
	ReservedNames    []string `yaml:"reserved_names"`     // Reserved field names

	CustomImports []string `yaml:"custom_imports"` // Additional proto imports

	// Additional file options (e.g., csharp_namespace, php_namespace, ruby_package, etc.)
	// These will be written as "option <key> = "<value>";" in the generated proto file
	Options map[string]string `yaml:"options"`

	// Stub generation configuration
	GenerateStubs *StubConfig `yaml:"generate_stubs,omitempty"`

	// Embed common configuration that can be overridden at plugin level
	goschemagen.Config `yaml:",inline"`
}

// StubConfig configures stub generation for type preservation
type StubConfig struct {
	Enabled                  bool              `yaml:"enabled"`
	AdapterPackage           string            `yaml:"adapter_package"`
	OutputDir                string            `yaml:"output_dir"` // Directory for generated protobuf Go files (default: "v1")
	OriginalServiceInterface bool              `yaml:"original_service_interface"`
	TypeMappings             TypeMappingConfig `yaml:"type_mappings"`
	StreamingSupport         bool              `yaml:"streaming_support"`
	RegistrationHelpers      bool              `yaml:"registration_helpers"`
	Templates                TemplateConfig    `yaml:"templates"`
}

// TypeMappingConfig configures how original types map to protobuf types
type TypeMappingConfig struct {
	AutoDetect        bool              `yaml:"auto_detect"`
	PreserveTimeTypes bool              `yaml:"preserve_time_types"`
	CustomMappings    map[string]string `yaml:"custom_mappings"`
}

func NewConfig() *Config {
	// Parse the nested configuration structure
	type ConfigFile struct {
		Plugins struct {
			Protobuf Config `yaml:"protobuf"`
		} `yaml:"plugins"`
	}

	configFile := &ConfigFile{}
	err := yaml.Unmarshal(DefaultConfigData, configFile)
	if err != nil {
		panic("failed to parse default protobuf config: " + err.Error())
	}
	return &configFile.Plugins.Protobuf
}
