package plugin

import (
	"strings"

	"github.com/pablor21/gonnotation/annotations"
	"github.com/pablor21/gonnotation/parser"
	"github.com/pablor21/goschemagen"
	"github.com/pablor21/protoschemagen/spec"
	"gopkg.in/yaml.v3"
)

// Plugin implements parser.Plugin for Protobuf schema generation
type Plugin struct {
	config *Config
}

func NewPlugin(config *Config) *Plugin {
	if config == nil {
		config = NewConfig()
	}
	return &Plugin{
		config: config,
	}
}

// GetConfig returns the plugin configuration
func (p *Plugin) GetConfig() *Config {
	return p.config
}

func (p *Plugin) Name() string {
	return "protobuf"
}

func (p *Plugin) Specs() []string {
	return []string{"protobuf", "proto", "grpc"}
}

func (p *Plugin) Definitions() annotations.PluginDefinitions {
	return spec.Specs
}

func (p *Plugin) FileExtension() string {
	// Extract extension from OutputFileName pattern
	// e.g., "{schema_name}.proto" -> ".proto"
	if idx := strings.LastIndex(p.config.OutputFileName, "."); idx != -1 {
		return p.config.OutputFileName[idx:]
	}
	return ".proto"
}

func (p *Plugin) AcceptsAnnotation(name string) bool {
	if name == "" {
		return false
	}
	n := strings.ToLower(name)

	// Accept prefixed compact forms (protoMessage) or dot forms (proto.message)
	if strings.HasPrefix(n, "proto") || strings.HasPrefix(n, "protobuf") || strings.HasPrefix(n, "grpc") {
		return true
	}
	// Accept canonical (prefix-free) annotation names defined in specs
	for _, ann := range spec.Specs.Annotations {
		if n == strings.ToLower(ann.Name) {
			return true
		}
		for _, alias := range ann.Aliases {
			if n == strings.ToLower(alias) {
				return true
			}
		}
	}

	return false
}

func (p *Plugin) ValidateConfig(config any) error {
	// Validate Protobuf-specific config
	return nil
}

func (p *Plugin) Generate(ctx *parser.GenerationContext) ([]byte, error) {
	// Use GenerateMulti and return first file for backward compatibility
	output, err := p.GenerateMulti(ctx)
	if err != nil {
		return nil, err
	}
	if len(output.Files) == 0 {
		return []byte{}, nil
	}
	return output.Files[0].Content, nil
}

func (p *Plugin) GenerateMulti(ctx *parser.GenerationContext) (*parser.GeneratedOutput, error) {
	// Parse format generator config from context
	cfg := p.config
	if cfg == nil {
		cfg = NewConfig()
	}

	// If format generator config is provided in context, unmarshal it
	if ctx.PluginConfig != nil {
		// Convert to YAML and back to parse into typed Config
		data, err := yaml.Marshal(ctx.PluginConfig)
		if err == nil {
			_ = yaml.Unmarshal(data, cfg) // Ignore error, will use defaults
		}
	}

	// Support backward compatibility: use deprecated fields if new ones are empty
	if cfg.GenerationStrategy == "" && cfg.GenStrategy != "" {
		cfg.GenerationStrategy = cfg.GenStrategy
	}
	if cfg.GenerationStrategy == "" {
		cfg.GenerationStrategy = parser.GenStrategySingle
	}

	// Update format generator config
	p.config = cfg

	// Create generator with context
	genContext := &goschemagen.GenerationContext{
		GenerationContext: *ctx, // Copy the context content
	}

	gen := &Generator{
		formatGen: p,
		ctx:       genContext,
	}

	return gen.GenerateMulti()
}
