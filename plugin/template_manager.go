package plugin

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var embeddedTemplates embed.FS

// TemplateSource represents the source of templates
type TemplateSource interface {
	LoadTemplate(name string) (string, error)
	ListTemplates() ([]string, error)
}

// EmbeddedTemplateSource loads templates from embedded filesystem
type EmbeddedTemplateSource struct{}

// LoadTemplate loads a template from the embedded filesystem
func (e *EmbeddedTemplateSource) LoadTemplate(name string) (string, error) {
	templatePath := fmt.Sprintf("templates/%s.go.tmpl", name)
	content, err := embeddedTemplates.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded template %s: %w", name, err)
	}
	return string(content), nil
}

// ListTemplates returns all available template names
func (e *EmbeddedTemplateSource) ListTemplates() ([]string, error) {
	var templates []string
	err := fs.WalkDir(embeddedTemplates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go.tmpl") {
			// Extract template name (remove templates/ prefix and .go.tmpl suffix)
			name := strings.TrimPrefix(path, "templates/")
			name = strings.TrimSuffix(name, ".go.tmpl")
			templates = append(templates, name)
		}
		return nil
	})
	return templates, err
}

// FileTemplateSource loads templates from filesystem
type FileTemplateSource struct {
	BasePath string
}

// LoadTemplate loads a template from the filesystem
func (f *FileTemplateSource) LoadTemplate(name string) (string, error) {
	templatePath := filepath.Join(f.BasePath, name+".go.tmpl")
	content, err := fs.ReadFile(nil, templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}
	return string(content), nil
}

// ListTemplates returns all available template names from filesystem
func (f *FileTemplateSource) ListTemplates() ([]string, error) {
	var templates []string
	err := filepath.WalkDir(f.BasePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go.tmpl") {
			// Extract template name relative to base path
			relPath, err := filepath.Rel(f.BasePath, path)
			if err != nil {
				return err
			}
			name := strings.TrimSuffix(relPath, ".go.tmpl")
			templates = append(templates, name)
		}
		return nil
	})
	return templates, err
}

// TemplateManager manages template loading and caching
type TemplateManager struct {
	source TemplateSource
	cache  map[string]*template.Template
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(source TemplateSource) *TemplateManager {
	return &TemplateManager{
		source: source,
		cache:  make(map[string]*template.Template),
	}
}

// GetTemplate loads and caches a template
func (tm *TemplateManager) GetTemplate(name string) (*template.Template, error) {
	// Check cache first
	if tmpl, exists := tm.cache[name]; exists {
		return tmpl, nil
	}

	// Load from source
	content, err := tm.source.LoadTemplate(name)
	if err != nil {
		return nil, err
	}

	// Parse template
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	// Cache it
	tm.cache[name] = tmpl
	return tmpl, nil
}

// TemplateConfig holds configuration for configurable templates
type TemplateConfig struct {
	// Template source configuration
	TemplateSource   string `yaml:"template_source"`    // "embedded" or "filesystem"
	TemplateBasePath string `yaml:"template_base_path"` // Base path for filesystem templates

	// Template name overrides (allows customizing which template file to use)
	TypesTemplate        string `yaml:"types_template"`
	ServiceTemplate      string `yaml:"service_template"`
	AdapterTemplate      string `yaml:"adapter_template"`
	ClientTemplate       string `yaml:"client_template"`
	BridgeTemplate       string `yaml:"bridge_template"`
	RegistrationTemplate string `yaml:"registration_template"`

	// Import configurations
	ModulePath      string `yaml:"module_path"`      // Base module path
	ProtobufPackage string `yaml:"protobuf_package"` // Generated protobuf package
	ProtobufAlias   string `yaml:"protobuf_alias"`   // Alias for protobuf package (default: pb)
}

// GetDefaultTemplateConfig returns default template configuration using embedded templates
func GetDefaultTemplateConfig() *TemplateConfig {
	return &TemplateConfig{
		TemplateSource:       "embedded",
		TypesTemplate:        "types",
		ServiceTemplate:      "service",
		AdapterTemplate:      "adapter",
		ClientTemplate:       "client",
		BridgeTemplate:       "bridge",
		RegistrationTemplate: "registration",
		ModulePath:           "", // Will be detected from generation context
		ProtobufPackage:      "", // Will be detected from options.go_package
		ProtobufAlias:        "pb",
	}
}

// CreateTemplateManager creates a template manager based on configuration
func (config *TemplateConfig) CreateTemplateManager() (*TemplateManager, error) {
	var source TemplateSource

	switch config.TemplateSource {
	case "embedded", "":
		source = &EmbeddedTemplateSource{}
	case "filesystem":
		if config.TemplateBasePath == "" {
			return nil, fmt.Errorf("template_base_path is required when using filesystem template source")
		}
		source = &FileTemplateSource{BasePath: config.TemplateBasePath}
	default:
		return nil, fmt.Errorf("unsupported template source: %s", config.TemplateSource)
	}

	return NewTemplateManager(source), nil
}

// GetTemplateNames returns the template names for each file type
func (config *TemplateConfig) GetTemplateNames() map[string]string {
	return map[string]string{
		"types":        config.TypesTemplate,
		"service":      config.ServiceTemplate,
		"adapter":      config.AdapterTemplate,
		"client":       config.ClientTemplate,
		"bridge":       config.BridgeTemplate,
		"registration": config.RegistrationTemplate,
	}
}
