package plugin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pablor21/gonnotation/parser"
	"github.com/pablor21/goschemagen"
)

// GenerateMulti generates multiple proto files based on generation strategy
func (g *Generator) GenerateMulti() (*parser.GeneratedOutput, error) {
	strategy := g.formatGen.config.GenerationStrategy

	var output *parser.GeneratedOutput
	var err error

	switch strategy {
	case parser.GenStrategySingle:
		output, err = g.generateSingle()
	case parser.GenStrategyFollow:
		output, err = g.generateFollow()
	case parser.GenStrategyPackage:
		output, err = g.generatePackage()
	case parser.GenStrategyNamespace:
		output, err = g.generateNamespace()
	default:
		output, err = g.generateSingle()
	}

	if err != nil {
		return nil, err
	}

	// Generate stubs if enabled (but skip protoc for now)
	if g.formatGen.config.GenerateStubs != nil && g.formatGen.config.GenerateStubs.Enabled {
		stubGen, err := NewStubGenerator(g.formatGen.config.GenerateStubs, g.formatGen.config, g.ctx)
		if err != nil {
			g.ctx.Logger.Info(fmt.Sprintf("Failed to create stub generator: %v", err))
		} else {
			// Generate only the adapter files, skip protoc (will be done later)
			if err := stubGen.generateAdapterFilesOnly(); err != nil {
				g.ctx.Logger.Info(fmt.Sprintf("Failed to generate stubs: %v", err))
			} else {
				g.ctx.Logger.Info("Generated type-preserving stubs successfully")
			}
		}
	}

	return output, nil
}

// generateSingle generates a single proto file with all schemas
func (g *Generator) generateSingle() (*parser.GeneratedOutput, error) {
	content, err := g.Generate()
	if err != nil {
		return nil, err
	}

	filename := g.resolveFileName("schema", "schema")

	return &parser.GeneratedOutput{
		Files: []*parser.GeneratedFile{
			{
				Path:    filename,
				Content: content,
			},
		},
		IsSingleFile: true,
	}, nil
}

// generateFollow generates one proto file per Go source file
func (g *Generator) generateFollow() (*parser.GeneratedOutput, error) {
	fileGroups := make(map[string]*fileGroup)

	// Group types by source file
	for _, s := range g.ctx.Structs {
		if s.SourceFile == "" {
			continue
		}

		group := fileGroups[s.SourceFile]
		if group == nil {
			group = &fileGroup{
				sourceFile: s.SourceFile,
				structs:    []*parser.StructInfo{},
				enums:      []*parser.EnumInfo{},
				interfaces: []*parser.InterfaceInfo{},
			}
			fileGroups[s.SourceFile] = group
		}
		group.structs = append(group.structs, s)
	}

	for _, e := range g.ctx.Enums {
		if e.SourceFile == "" {
			continue
		}

		group := fileGroups[e.SourceFile]
		if group == nil {
			group = &fileGroup{
				sourceFile: e.SourceFile,
				structs:    []*parser.StructInfo{},
				enums:      []*parser.EnumInfo{},
				interfaces: []*parser.InterfaceInfo{},
			}
			fileGroups[e.SourceFile] = group
		}
		group.enums = append(group.enums, e)
	}

	for _, i := range g.ctx.Interfaces {
		if i.SourceFile == "" {
			continue
		}

		group := fileGroups[i.SourceFile]
		if group == nil {
			group = &fileGroup{
				sourceFile: i.SourceFile,
				structs:    []*parser.StructInfo{},
				enums:      []*parser.EnumInfo{},
				interfaces: []*parser.InterfaceInfo{},
			}
			fileGroups[i.SourceFile] = group
		}
		group.interfaces = append(group.interfaces, i)
	}

	// First pass: populate file type mappings
	for sourceFile, group := range fileGroups {
		baseName := filepath.Base(sourceFile)
		baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		filename := g.resolveFileName(baseName, baseName)

		// Add file mapping to context
		mapping := g.ctx.AddFileTypeMapping(filename)

		// Add types to mapping
		for _, s := range group.structs {
			mapping.Structs[s.Name] = true
		}
		for _, e := range group.enums {
			mapping.Enums[e.Name] = true
		}
		for _, i := range group.interfaces {
			mapping.Services[i.Name] = true
		}
	}

	// Second pass: generate files with proper imports
	var files []*parser.GeneratedFile

	for sourceFile, group := range fileGroups {
		// Use the base name of the source file without extension
		baseName := filepath.Base(sourceFile)
		baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		filename := g.resolveFileName(baseName, baseName)

		// Set current file context for imports
		g.currentFile = filename

		content, err := g.generateForGroup(group)
		if err != nil {
			return nil, fmt.Errorf("error generating proto for %s: %w", sourceFile, err)
		}

		if len(content) == 0 {
			continue
		}

		files = append(files, &parser.GeneratedFile{
			Path:    filename,
			Content: content,
			Metadata: map[string]any{
				"source_file": sourceFile,
			},
		})
	}

	return &parser.GeneratedOutput{
		Files:        files,
		IsSingleFile: false,
	}, nil
}

// generatePackage generates one proto file per Go package
func (g *Generator) generatePackage() (*parser.GeneratedOutput, error) {
	packageGroups := make(map[string]*fileGroup)

	// Group types by package
	for _, s := range g.ctx.Structs {
		if s.Package == "" {
			continue
		}

		group := packageGroups[s.Package]
		if group == nil {
			group = &fileGroup{
				packageName: s.Package,
				structs:     []*parser.StructInfo{},
				enums:       []*parser.EnumInfo{},
				interfaces:  []*parser.InterfaceInfo{},
			}
			packageGroups[s.Package] = group
		}
		group.structs = append(group.structs, s)
	}

	for _, e := range g.ctx.Enums {
		if e.Package == "" {
			continue
		}

		group := packageGroups[e.Package]
		if group == nil {
			group = &fileGroup{
				packageName: e.Package,
				structs:     []*parser.StructInfo{},
				enums:       []*parser.EnumInfo{},
				interfaces:  []*parser.InterfaceInfo{},
			}
			packageGroups[e.Package] = group
		}
		group.enums = append(group.enums, e)
	}

	for _, i := range g.ctx.Interfaces {
		if i.Package == "" {
			continue
		}

		group := packageGroups[i.Package]
		if group == nil {
			group = &fileGroup{
				packageName: i.Package,
				structs:     []*parser.StructInfo{},
				enums:       []*parser.EnumInfo{},
				interfaces:  []*parser.InterfaceInfo{},
			}
			packageGroups[i.Package] = group
		}
		group.interfaces = append(group.interfaces, i)
	}

	// Generate a file for each package
	var files []*parser.GeneratedFile

	for pkgName, group := range packageGroups {
		content, err := g.generateForGroup(group)
		if err != nil {
			return nil, fmt.Errorf("error generating proto for package %s: %w", pkgName, err)
		}

		if len(content) == 0 {
			continue
		}

		// Use the last part of the package path as the filename
		parts := strings.Split(pkgName, "/")
		baseName := parts[len(parts)-1]

		filename := g.resolveFileName(baseName, baseName)

		files = append(files, &parser.GeneratedFile{
			Path:    filename,
			Content: content,
			Metadata: map[string]any{
				"package": pkgName,
			},
		})
	}

	return &parser.GeneratedOutput{
		Files:        files,
		IsSingleFile: false,
	}, nil
}

// generateNamespace generates one .proto file per namespace
func (g *Generator) generateNamespace() (*parser.GeneratedOutput, error) {
	namespaceGroups := make(map[string]*fileGroup)

	// Group types by namespace
	for _, s := range g.ctx.Structs {
		namespace := s.Namespace
		if namespace == "" {
			namespace = "default"
		}

		group := namespaceGroups[namespace]
		if group == nil {
			group = &fileGroup{
				namespace:  namespace,
				structs:    []*parser.StructInfo{},
				enums:      []*parser.EnumInfo{},
				interfaces: []*parser.InterfaceInfo{},
			}
			namespaceGroups[namespace] = group
		}
		group.structs = append(group.structs, s)
	}

	for _, e := range g.ctx.Enums {
		namespace := e.Namespace
		if namespace == "" {
			namespace = "default"
		}

		group := namespaceGroups[namespace]
		if group == nil {
			group = &fileGroup{
				namespace:  namespace,
				structs:    []*parser.StructInfo{},
				enums:      []*parser.EnumInfo{},
				interfaces: []*parser.InterfaceInfo{},
			}
			namespaceGroups[namespace] = group
		}
		group.enums = append(group.enums, e)
	}

	// Interfaces are not typically used in protobuf, but include for completeness
	for _, i := range g.ctx.Interfaces {
		namespace := i.Namespace
		if namespace == "" {
			namespace = "default"
		}

		group := namespaceGroups[namespace]
		if group == nil {
			group = &fileGroup{
				namespace:  namespace,
				structs:    []*parser.StructInfo{},
				enums:      []*parser.EnumInfo{},
				interfaces: []*parser.InterfaceInfo{},
			}
			namespaceGroups[namespace] = group
		}
		group.interfaces = append(group.interfaces, i)
	}

	// Generate a file for each namespace
	var files []*parser.GeneratedFile

	for namespaceName, group := range namespaceGroups {
		content, err := g.generateForGroup(group)
		if err != nil {
			return nil, fmt.Errorf("error generating proto for namespace %s: %w", namespaceName, err)
		}

		if len(content) == 0 {
			continue
		}

		filename := g.resolveFileName(namespaceName, namespaceName)

		files = append(files, &parser.GeneratedFile{
			Path:    filename,
			Content: content,
			Metadata: map[string]any{
				"namespace": namespaceName,
			},
		})
	}

	return &parser.GeneratedOutput{
		Files:        files,
		IsSingleFile: false,
	}, nil
}

// fileGroup represents a group of types to be generated together
type fileGroup struct {
	sourceFile  string
	packageName string
	namespace   string
	structs     []*parser.StructInfo
	enums       []*parser.EnumInfo
	interfaces  []*parser.InterfaceInfo
}

// generateForGroup generates protobuf schema for a specific group of types
func (g *Generator) generateForGroup(group *fileGroup) ([]byte, error) {
	// Create a temporary context with only the types in this group
	tempCtx := &goschemagen.GenerationContext{
		GenerationContext: parser.GenerationContext{
			Structs:          group.structs,
			Enums:            group.enums,
			Interfaces:       group.interfaces,
			Functions:        g.ctx.Functions,    // Include functions for struct services
			AllFunctions:     g.ctx.AllFunctions, // Include all functions for reference
			CoreConfig:       g.ctx.CoreConfig,
			Logger:           g.ctx.Logger,
			FileTypeMappings: g.ctx.FileTypeMappings, // Pass through file mappings
		},
	}

	// Create a temporary generator with the filtered context
	tempGen := &Generator{
		formatGen:     g.formatGen,
		ctx:           tempCtx,
		fieldNumbers:  make(map[string]int),
		currentNumber: g.formatGen.config.StartFieldNumber,
		currentFile:   g.currentFile, // Pass current file context
	}

	// Generate the schema
	return tempGen.Generate()
}

// resolveFileName resolves the output filename from the pattern
func (g *Generator) resolveFileName(schemaName, name string) string {
	// Use Output field if set, otherwise fall back to OutputFileName
	pattern := g.formatGen.config.Output
	if pattern == "" {
		pattern = g.formatGen.config.OutputFileName
	}
	if pattern == "" {
		pattern = "{schema_name}.proto"
	}

	// Replace placeholders
	result := strings.ReplaceAll(pattern, "{schema_name}", schemaName)
	result = strings.ReplaceAll(result, "{name}", name)

	return result
}
