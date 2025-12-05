package plugin

import (
	"fmt"
	"go/ast"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pablor21/gonnotation/annotations"
	"github.com/pablor21/gonnotation/parser"
	"github.com/pablor21/goschemagen"
)

// StubGenerator generates type adapters and original-type service interfaces
type StubGenerator struct {
	config          *StubConfig
	pluginConfig    *Config
	ctx             *goschemagen.GenerationContext
	originalTypes   map[string]*TypeInfo
	protoTypes      map[string]*TypeInfo
	services        []*ServiceInfo
	templateManager *TemplateManager

	// Reference to main generator for parsed data
	mainGenerator *Generator
}

// TypeInfo holds information about a type for mapping
type TypeInfo struct {
	Name        string
	Package     string
	FullName    string
	Fields      []*FieldInfo
	IsMessage   bool
	IsEnum      bool
	IsService   bool
	Annotations []annotations.Annotation
}

// FieldInfo holds information about struct fields
type FieldInfo struct {
	Name         string // Field name from gonnotation (might be different from GoName)
	GoName       string // Go field name from gonnotation
	Type         string
	Tag          string // Struct tag from gonnotation
	ProtoName    string
	ProtoNumber  int
	ProtoType    string
	IsOptional   bool
	IsRepeated   bool
	IsMap        bool
	MapKeyType   string
	MapValueType string
	IsEmbedded   bool // If this is an embedded field
}

// ServiceInfo holds information about service interfaces
type ServiceInfo struct {
	Name     string
	Package  string
	FullName string
	Methods  []*MethodInfo
	IsStruct bool // true if it's a concrete struct, false if it's an interface
}

// MethodInfo holds information about service methods
type MethodInfo struct {
	Name               string
	InputType          string // Protobuf type (e.g., "google.protobuf.Int64Value")
	OutputType         string // Protobuf type (e.g., "User")
	OriginalInputType  string // Original Go type (e.g., "int64")
	OriginalOutputType string // Original Go type (e.g., "*User")
	IsStreaming        bool
	ClientStream       bool
	ServerStream       bool
	HasContext         bool // Whether original method has context.Context parameter
}

// NewStubGenerator creates a new stub generator
func NewStubGenerator(config *StubConfig, pluginConfig *Config, ctx *goschemagen.GenerationContext, mainGen *Generator) (*StubGenerator, error) {
	generator := &StubGenerator{
		config:        config,
		pluginConfig:  pluginConfig,
		ctx:           ctx,
		originalTypes: make(map[string]*TypeInfo),
		protoTypes:    make(map[string]*TypeInfo),
		services:      make([]*ServiceInfo, 0),
		mainGenerator: mainGen,
	}

	// Initialize template manager
	templateConfig := GetDefaultTemplateConfig()
	templateManager, err := templateConfig.CreateTemplateManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize template manager: %w", err)
	}
	generator.templateManager = templateManager

	return generator, nil
}

// Generate generates stub files with type adapters
func (g *StubGenerator) Generate() error {
	if !g.config.Enabled {
		return nil
	}

	// Step 1: Analyze original types from parsed context
	if err := g.analyzeOriginalTypes(); err != nil {
		return fmt.Errorf("failed to analyze original types: %w", err)
	}

	// Step 2: Generate type adapters
	if err := g.generateTypeAdapters(); err != nil {
		return fmt.Errorf("failed to generate type adapters: %w", err)
	}

	// Step 3: Generate original-type service interface
	if g.config.OriginalServiceInterface {
		if err := g.generateOriginalServiceInterface(); err != nil {
			return fmt.Errorf("failed to generate original service interface: %w", err)
		}
	}

	// Step 4: Generate service adapter
	if err := g.generateServiceAdapter(); err != nil {
		return fmt.Errorf("failed to generate service adapter: %w", err)
	}

	// Step 5: Generate client
	if err := g.generateClient(); err != nil {
		return fmt.Errorf("failed to generate client: %w", err)
	}

	// Step 6: Generate service bridge
	if err := g.generateServiceBridge(); err != nil {
		return fmt.Errorf("failed to generate service bridge: %w", err)
	}

	// Step 7: Generate registration helpers
	if g.config.RegistrationHelpers {
		if err := g.generateRegistrationHelpers(); err != nil {
			return fmt.Errorf("failed to generate registration helpers: %w", err)
		}
	}

	// Step 8: Generate protobuf Go files automatically
	if err := g.generateProtobufGoFiles(); err != nil {
		return fmt.Errorf("failed to generate protobuf Go files: %w", err)
	}

	return nil
}

// generateAdapterFilesOnly generates only the adapter files without running protoc
func (g *StubGenerator) generateAdapterFilesOnly() error {
	if !g.config.Enabled {
		return nil
	}

	// Step 1: Analyze original types from parsed context
	if err := g.analyzeOriginalTypes(); err != nil {
		return fmt.Errorf("failed to analyze original types: %w", err)
	}

	// Step 2: Generate type adapters
	if err := g.generateTypeAdapters(); err != nil {
		return fmt.Errorf("failed to generate type adapters: %w", err)
	}

	// Step 3: Generate original-type service interface
	if g.config.OriginalServiceInterface {
		if err := g.generateOriginalServiceInterface(); err != nil {
			return fmt.Errorf("failed to generate original service interface: %w", err)
		}
	}

	// Step 4: Generate service adapter
	if err := g.generateServiceAdapter(); err != nil {
		return fmt.Errorf("failed to generate service adapter: %w", err)
	}

	// Step 5: Generate client
	if err := g.generateClient(); err != nil {
		return fmt.Errorf("failed to generate client: %w", err)
	}

	// Step 6: Generate service bridge
	if err := g.generateServiceBridge(); err != nil {
		return fmt.Errorf("failed to generate service bridge: %w", err)
	}

	// Step 7: Generate registration helpers
	if g.config.RegistrationHelpers {
		if err := g.generateRegistrationHelpers(); err != nil {
			return fmt.Errorf("failed to generate registration helpers: %w", err)
		}
	}

	// Skip protoc generation - will be done later
	return nil
}

// analyzeOriginalTypes scans the parsed context for types with protobuf annotations
func (g *StubGenerator) analyzeOriginalTypes() error {
	// Include ALL structs from context - they were already filtered by the main generator
	for _, structInfo := range g.ctx.Structs {
		// Skip service structs - they don't need conversion functions
		isService := false
		for _, ann := range structInfo.Annotations {
			if g.isServiceAnnotation(&ann) {
				isService = true
				break
			}
		}
		if isService {
			continue
		}

		// Skip generic types (they can't be converted to protobuf)
		if strings.Contains(structInfo.Name, "[") || len(structInfo.Annotations) == 0 {
			g.ctx.Logger.Debug(fmt.Sprintf("Skipping type %s (generic or no annotations)", structInfo.Name))
			continue
		}

		// Include ALL non-service, non-generic structs
		typeInfo := &TypeInfo{
			Name:        structInfo.Name,
			Package:     structInfo.PackagePath,
			FullName:    fmt.Sprintf("%s.%s", structInfo.PackagePath, structInfo.Name),
			IsMessage:   true,
			Annotations: structInfo.Annotations,
			Fields:      make([]*FieldInfo, 0),
		}

		// Analyze fields
		for _, field := range structInfo.Fields {
			fieldInfo := g.analyzeField(field)
			typeInfo.Fields = append(typeInfo.Fields, fieldInfo)
		}

		g.originalTypes[typeInfo.Name] = typeInfo
		g.ctx.Logger.Debug(fmt.Sprintf("Added type for conversion: %s", typeInfo.Name))
	} // Add enums from context
	for _, enumInfo := range g.ctx.Enums {
		typeInfo := &TypeInfo{
			Name:        enumInfo.Name,
			Package:     enumInfo.PackagePath,
			FullName:    fmt.Sprintf("%s.%s", enumInfo.PackagePath, enumInfo.Name),
			IsMessage:   false,
			IsEnum:      true, // Mark as enum
			Annotations: enumInfo.Annotations,
			Fields:      make([]*FieldInfo, 0),
		}

		g.originalTypes[typeInfo.Name] = typeInfo
		g.ctx.Logger.Debug(fmt.Sprintf("Added enum for conversion: %s", typeInfo.Name))
	}

	// Use pre-parsed services from main generator instead of re-parsing
	parsedServices := g.mainGenerator.GetParsedServices()
	for _, protoService := range parsedServices {
		serviceInfo := &ServiceInfo{
			Name:     protoService.Name,
			Package:  "", // Will be filled from context
			FullName: protoService.Name,
			Methods:  make([]*MethodInfo, 0),
			IsStruct: protoService.IsStruct,
		}

		// Convert proto methods to stub methods
		for _, protoMethod := range protoService.Methods {
			methodInfo := &MethodInfo{
				Name:               protoMethod.Name,
				InputType:          protoMethod.InputType,
				OutputType:         protoMethod.OutputType,
				OriginalInputType:  protoMethod.OriginalInputType,
				OriginalOutputType: protoMethod.OriginalOutputType,
				IsStreaming:        protoMethod.IsStreaming,
				ClientStream:       protoMethod.ClientStream,
				ServerStream:       protoMethod.ServerStream,
				HasContext:         protoMethod.HasContext,
			}
			serviceInfo.Methods = append(serviceInfo.Methods, methodInfo)
		}

		g.services = append(g.services, serviceInfo)
	}

	// Analyze enums
	for _, enumInfo := range g.ctx.Enums {
		for _, ann := range enumInfo.Annotations {
			if g.isEnumAnnotation(&ann) {
				typeInfo := &TypeInfo{
					Name:        enumInfo.Name,
					Package:     enumInfo.PackagePath,
					FullName:    fmt.Sprintf("%s.%s", enumInfo.PackagePath, enumInfo.Name),
					IsEnum:      true,
					Annotations: enumInfo.Annotations,
					Fields:      make([]*FieldInfo, 0), // Enums don't have fields
				}

				g.originalTypes[typeInfo.Name] = typeInfo
				break
			}
		}
	}

	g.ctx.Logger.Debug(fmt.Sprintf("Analysis complete: %d services detected", len(g.services)))
	for _, service := range g.services {
		g.ctx.Logger.Debug(fmt.Sprintf("  Service %s (%s) with %d methods", service.Name, map[bool]string{true: "struct", false: "interface"}[service.IsStruct], len(service.Methods)))
		for _, method := range service.Methods {
			g.ctx.Logger.Debug(fmt.Sprintf("    Method %s: %s -> %s", method.Name, method.InputType, method.OutputType))
		}
	}

	return nil
}

// analyzeField extracts field information for type mapping
func (g *StubGenerator) analyzeField(field *parser.FieldInfo) *FieldInfo {
	fieldInfo := &FieldInfo{
		Name:       field.Name,
		GoName:     field.GoName,
		Type:       g.typeExprToString(field.Type),
		IsEmbedded: field.IsEmbedded,
	}

	// Extract struct tag if available
	if field.Tag != nil {
		fieldInfo.Tag = field.Tag.Value
	}

	// Extract protobuf field information from annotations
	for _, ann := range field.Annotations {
		if g.isFieldAnnotation(&ann) {
			if num, ok := ann.Params["number"]; ok {
				fieldInfo.ProtoNumber = g.parseInt(num)
			}
			if name, ok := ann.Params["name"]; ok {
				fieldInfo.ProtoName = name
			} else {
				fieldInfo.ProtoName = g.toSnakeCase(field.Name)
			}
			if repeated, ok := ann.Params["repeated"]; ok {
				fieldInfo.IsRepeated = repeated == "true"
			}
			break
		}
	}

	// Determine protobuf type from Go type
	fieldInfo.ProtoType = g.goTypeToProtoType(fieldInfo.Type)

	return fieldInfo
}

// // addServiceMethods adds methods to a service from the Functions list
// // This works for both interface and struct services
// func (g *StubGenerator) addServiceMethods(serviceInfo *ServiceInfo, serviceName string) {
// 	if !serviceInfo.IsStruct {
// 		// For interface services, get methods from the interface definition
// 		for _, intfInfo := range g.ctx.Interfaces {
// 			if intfInfo.Name == serviceName {
// 				// If the interface itself has service annotation, include ALL methods
// 				hasServiceAnnotation := false
// 				for _, ann := range intfInfo.Annotations {
// 					name := strings.ToLower(ann.Name)
// 					if name == "proto.service" || name == "service" || strings.HasSuffix(name, ".service") {
// 						hasServiceAnnotation = true
// 						break
// 					}
// 				}

// 				for _, method := range intfInfo.Methods {
// 					// Include method if interface has service annotation OR method has individual RPC annotation
// 					if hasServiceAnnotation || g.hasRPCAnnotationForMethod(method) {
// 						methodInfo := g.createMethodInfoFromInterfaceMethod(method)
// 						serviceInfo.Methods = append(serviceInfo.Methods, methodInfo)
// 					}
// 				}
// 				break
// 			}
// 		}
// 		return
// 	}

// 	// For struct services, find all functions that belong to this service
// 	g.ctx.Logger.Debug(fmt.Sprintf("Looking for methods for struct service: %s", serviceName))
// 	g.ctx.Logger.Debug(fmt.Sprintf("Total functions in context: %d", len(g.ctx.Functions)))

// 	for _, fn := range g.ctx.Functions {
// 		// Check if this function is a method of our service
// 		isServiceMethod := false

// 		if fn.Receiver != nil {
// 			g.ctx.Logger.Debug(fmt.Sprintf("Function %s has receiver: %s", fn.Name, fn.Receiver.TypeName))
// 			// For struct services, check receiver type
// 			if fn.Receiver.TypeName == serviceName {
// 				isServiceMethod = true
// 				g.ctx.Logger.Debug(fmt.Sprintf("Function %s matches service %s", fn.Name, serviceName))
// 			}
// 		} else {
// 			g.ctx.Logger.Debug(fmt.Sprintf("Function %s has no receiver", fn.Name))
// 		}

// 		// Skip if not a service method or doesn't have @rpc annotation
// 		if !isServiceMethod {
// 			g.ctx.Logger.Debug(fmt.Sprintf("Function %s is not a service method", fn.Name))
// 			continue
// 		}

// 		// Check RPC annotation
// 		hasRPC := g.hasRPCAnnotation(fn)
// 		g.ctx.Logger.Debug(fmt.Sprintf("Function %s has %d annotations, RPC annotation: %v", fn.Name, len(fn.Annotations), hasRPC))
// 		for i, ann := range fn.Annotations {
// 			g.ctx.Logger.Debug(fmt.Sprintf("  Annotation %d: %s", i, ann.Name))
// 		}

// 		if !hasRPC {
// 			g.ctx.Logger.Debug(fmt.Sprintf("Function %s has no RPC annotation", fn.Name))
// 			continue
// 		}

// 		// Create method info from function
// 		methodInfo := g.createMethodInfoFromFunction(fn)
// 		serviceInfo.Methods = append(serviceInfo.Methods, methodInfo)
// 		g.ctx.Logger.Debug(fmt.Sprintf("Added method %s to service %s", methodInfo.Name, serviceName))
// 	}
// }

// // hasRPCAnnotation checks if a function has @rpc annotation
// func (g *StubGenerator) hasRPCAnnotation(fn *parser.FunctionInfo) bool {
// 	for _, ann := range fn.Annotations {
// 		name := strings.ToLower(ann.Name)
// 		if name == "rpc" || strings.HasSuffix(name, ".rpc") {
// 			return true
// 		}
// 	}
// 	return false
// }

// // hasRPCAnnotationForMethod checks if an interface method has @rpc annotation
// func (g *StubGenerator) hasRPCAnnotationForMethod(method *parser.MethodInfo) bool {
// 	for _, ann := range method.Annotations {
// 		name := strings.ToLower(ann.Name)
// 		if name == "rpc" || strings.HasSuffix(name, ".rpc") {
// 			return true
// 		}
// 	}
// 	return false
// }

// // createMethodInfoFromInterfaceMethod creates MethodInfo from interface method
// func (g *StubGenerator) createMethodInfoFromInterfaceMethod(method *parser.MethodInfo) *MethodInfo {
// 	methodInfo := &MethodInfo{
// 		Name: method.Name,
// 	}

// 	// Extract RPC information from annotations
// 	for _, ann := range method.Annotations {
// 		if g.isRPCAnnotation(&ann) {
// 			if input, ok := ann.Params["input"]; ok {
// 				methodInfo.InputType = input
// 			}
// 			if output, ok := ann.Params["output"]; ok {
// 				methodInfo.OutputType = output
// 			}
// 			if clientStreaming, ok := ann.Params["client_streaming"]; ok {
// 				methodInfo.ClientStream = clientStreaming == "true"
// 			}
// 			if serverStreaming, ok := ann.Params["server_streaming"]; ok {
// 				methodInfo.ServerStream = serverStreaming == "true"
// 			}
// 			methodInfo.IsStreaming = methodInfo.ClientStream || methodInfo.ServerStream
// 			break
// 		}
// 	}

// 	return methodInfo
// }

// // createMethodInfoFromFunction creates MethodInfo from function
// func (g *StubGenerator) createMethodInfoFromFunction(fn *parser.FunctionInfo) *MethodInfo {
// 	methodInfo := &MethodInfo{
// 		Name: fn.Name,
// 	}

// 	// Extract input and output types from function signature
// 	// Input type (skip context.Context, take the actual request parameter)
// 	for i, param := range fn.Params {
// 		typeUtils := parser.NewTypeUtils()
// 		paramType := typeUtils.GetTypeName(param.Type)

// 		// Skip context.Context parameter (appears as just "Context" from GetTypeName)
// 		// Also skip the first parameter by default (usually context)
// 		if i == 0 || paramType == "Context" || strings.Contains(paramType, "context.Context") {
// 			continue
// 		}

// 		methodInfo.InputType = g.extractTypeFromParam(param)
// 		break
// 	}

// 	// Output type (first return value that's not error)
// 	for _, result := range fn.Results {
// 		typeUtils := parser.NewTypeUtils()
// 		resultType := typeUtils.GetTypeName(result.Type)
// 		if resultType != "error" {
// 			methodInfo.OutputType = g.extractTypeFromParam(result)
// 			break
// 		}
// 	}

// 	// Extract RPC information from annotations
// 	for _, ann := range fn.Annotations {
// 		if g.isRPCAnnotation(&ann) {
// 			if input, ok := ann.Params["input"]; ok {
// 				methodInfo.InputType = input
// 			}
// 			if output, ok := ann.Params["output"]; ok {
// 				methodInfo.OutputType = output
// 			}
// 			if clientStreaming, ok := ann.Params["client_streaming"]; ok {
// 				methodInfo.ClientStream = clientStreaming == "true"
// 			}
// 			if serverStreaming, ok := ann.Params["server_streaming"]; ok {
// 				methodInfo.ServerStream = serverStreaming == "true"
// 			}
// 			methodInfo.IsStreaming = methodInfo.ClientStream || methodInfo.ServerStream
// 			break
// 		}
// 	}

// 	return methodInfo
// }

// // extractTypeFromParam extracts the clean type name from a parameter
// func (g *StubGenerator) extractTypeFromParam(param *parser.ParamInfo) string {
// 	if param == nil {
// 		return ""
// 	}

// 	// Use TypeUtils to convert ast.Expr to string
// 	typeUtils := parser.NewTypeUtils()
// 	typeName := typeUtils.GetTypeName(param.Type)

// 	// Remove package prefix if it exists to get clean type name
// 	if dotIndex := strings.LastIndex(typeName, "."); dotIndex >= 0 {
// 		typeName = typeName[dotIndex+1:]
// 	}

// 	return typeName
// }

// Helper functions for annotation detection

func (g *StubGenerator) isServiceAnnotation(ann *annotations.Annotation) bool {
	name := strings.ToLower(ann.Name)
	return strings.Contains(name, "service") || name == "proto.service"
}

func (g *StubGenerator) isFieldAnnotation(ann *annotations.Annotation) bool {
	name := strings.ToLower(ann.Name)
	return strings.Contains(name, "field") || name == "proto.field"
}

func (g *StubGenerator) isEnumAnnotation(ann *annotations.Annotation) bool {
	name := strings.ToLower(ann.Name)
	return strings.Contains(name, "enum") || name == "proto.enum"
}

// func (g *StubGenerator) isRPCAnnotation(ann *annotations.Annotation) bool {
// 	name := strings.ToLower(ann.Name)
// 	return strings.Contains(name, "rpc") || name == "proto.rpc"
// }

// Helper functions
func (g *StubGenerator) parseInt(s string) int {
	// Simple integer parsing, could be enhanced
	if s == "" {
		return 0
	}
	// This is a simplified implementation
	// In real implementation, use strconv.Atoi with error handling
	return 1
}

func (g *StubGenerator) toSnakeCase(s string) string {
	// Special handling for common abbreviations
	switch s {
	case "ID":
		return "id"
	case "URL":
		return "url"
	case "API":
		return "api"
	case "HTTP":
		return "http"
	case "UUID":
		return "uuid"
	}

	// Simple snake_case conversion
	result := ""
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		result += strings.ToLower(string(r))
	}
	return result
}

func (g *StubGenerator) goTypeToProtoType(goType string) string {
	// Map Go types to protobuf types
	switch goType {
	case "string":
		return "string"
	case "int", "int32":
		return "int32"
	case "int64":
		return "int64"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "bool":
		return "bool"
	default:
		// For complex types, assume it's a message type
		return goType
	}
}

func (g *StubGenerator) typeExprToString(t interface{}) string {
	// Handle ast.Expr types properly
	if expr, ok := t.(ast.Expr); ok {
		return g.getGoTypeName(expr)
	}
	// fallback to string representation
	return fmt.Sprintf("%v", t)
}

// getGoTypeName extracts the type name from an AST expression
func (g *StubGenerator) getGoTypeName(t ast.Expr) string {
	switch v := t.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.StarExpr:
		return "*" + g.getGoTypeName(v.X)
	case *ast.ArrayType:
		return "[]" + g.getGoTypeName(v.Elt)
	case *ast.SelectorExpr:
		return g.getGoTypeName(v.X) + "." + v.Sel.Name
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", g.getGoTypeName(v.Key), g.getGoTypeName(v.Value))
	}
	return "unknown"
}

// getAdapterPackagePath returns the full package path for the adapter package
func (g *StubGenerator) getAdapterPackagePath() string {
	if g.config.AdapterPackage == "" {
		return "generated/adapter"
	}
	return g.config.AdapterPackage
}

// getTemplateConfig returns the template configuration, using defaults if not specified
func (g *StubGenerator) getTemplateConfig() *TemplateConfig {
	// Start with defaults
	templateConfig := GetDefaultTemplateConfig()

	// Auto-detect module path from context if not provided
	if templateConfig.ModulePath == "" && g.ctx != nil {
		// Try to get module path from the first parsed package
		for _, structInfo := range g.ctx.Structs {
			if structInfo.Package != "" {
				// Extract module path from full package path
				templateConfig.ModulePath = g.extractModulePath(structInfo.Package)
				break
			}
		}
	}

	// Auto-detect protobuf package from options.go_package
	if templateConfig.ProtobufPackage == "" && g.pluginConfig != nil && g.pluginConfig.Options != nil {
		if goPackage, exists := g.pluginConfig.Options["go_package"]; exists {
			templateConfig.ProtobufPackage = g.resolveProtobufPackage(goPackage)
		}
	}

	// Override with user configuration if provided
	if g.config.Templates.ModulePath != "" {
		templateConfig.ModulePath = g.config.Templates.ModulePath
	}
	if g.config.Templates.ProtobufPackage != "" {
		templateConfig.ProtobufPackage = g.config.Templates.ProtobufPackage
	}
	if g.config.Templates.ProtobufAlias != "" {
		templateConfig.ProtobufAlias = g.config.Templates.ProtobufAlias
	}
	if g.config.Templates.TypesTemplate != "" {
		templateConfig.TypesTemplate = g.config.Templates.TypesTemplate
	}
	if g.config.Templates.ServiceTemplate != "" {
		templateConfig.ServiceTemplate = g.config.Templates.ServiceTemplate
	}
	if g.config.Templates.AdapterTemplate != "" {
		templateConfig.AdapterTemplate = g.config.Templates.AdapterTemplate
	}
	if g.config.Templates.RegistrationTemplate != "" {
		templateConfig.RegistrationTemplate = g.config.Templates.RegistrationTemplate
	}

	return templateConfig
}

// writeFile writes content to a file in the adapter package
func (g *StubGenerator) writeFile(filename string, content []byte) error {
	// Skip writing files that are effectively empty (only package + imports)
	if g.isEffectivelyEmptyGoFile(content) {
		g.ctx.Logger.Debug(fmt.Sprintf("Skipping empty stub file: %s", filename))
		return nil
	}

	adapterDir := g.getAdapterPackagePath()

	// If adapter directory is not absolute, resolve it relative to config directory
	if !filepath.IsAbs(adapterDir) {
		if g.ctx.CoreConfig.ConfigDir != "" {
			adapterDir = filepath.Join(g.ctx.CoreConfig.ConfigDir, adapterDir)
		}
	}

	// Create adapter directory if it doesn't exist
	if err := os.MkdirAll(adapterDir, 0755); err != nil {
		return fmt.Errorf("failed to create adapter directory %s: %w", adapterDir, err)
	}

	fullPath := filepath.Join(adapterDir, filename)

	// Write content to file
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	g.ctx.Logger.Debug(fmt.Sprintf("Generated stub file: %s (%d bytes)", fullPath, len(content)))
	return nil
}

// isEffectivelyEmptyGoFile checks if a Go file contains only package declaration and imports
func (g *StubGenerator) isEffectivelyEmptyGoFile(content []byte) bool {
	contentStr := strings.TrimSpace(string(content))

	// If file is very small, likely just package + imports
	if len(contentStr) < 300 {
		lines := strings.Split(contentStr, "\n")
		hasActualCode := false

		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Skip empty lines, comments, package declaration, imports
			if line == "" ||
				strings.HasPrefix(line, "//") ||
				strings.HasPrefix(line, "package ") ||
				strings.HasPrefix(line, "import ") ||
				line == "import (" ||
				line == ")" ||
				strings.HasPrefix(line, "\t") && strings.Contains(line, "\"") {
				continue
			}
			// If we find any other line, it's actual code
			hasActualCode = true
			break
		}

		return !hasActualCode
	}

	return false
}

// generateProtobufGoFiles runs protoc to generate Go files from .proto schema(s)
func (g *StubGenerator) generateProtobufGoFiles() error {
	// Skip if protobuf package is not configured
	templateConfig := g.getTemplateConfig()
	if templateConfig.ProtobufPackage == "" {
		return nil
	}

	// Find all .proto files that were generated
	protoFiles, err := g.discoverProtoFiles()
	if err != nil {
		return fmt.Errorf("failed to discover proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		g.ctx.Logger.Debug("Skipping protobuf Go generation: no proto files found")
		return nil
	}

	// Extract the go_package option from config
	goPackage := templateConfig.ProtobufPackage
	if g.pluginConfig != nil && g.pluginConfig.Options != nil {
		if pkg, ok := g.pluginConfig.Options["go_package"]; ok && pkg != "" {
			goPackage = pkg
		}
	}

	// Determine output directory for protobuf Go files
	outputDir := g.config.OutputDir
	if outputDir == "" {
		// Default: use the last part of the go_package as the directory name
		outputDir = "v1" // fallback default
		if goPackage != "" {
			parts := strings.Split(goPackage, "/")
			if len(parts) > 0 {
				outputDir = parts[len(parts)-1] // use last segment (e.g., "v1" from "github.com/example/proto/starwars/v1")
			}
		}
	}

	// If output directory is not absolute, resolve it relative to config directory
	if !filepath.IsAbs(outputDir) {
		if g.ctx.CoreConfig.ConfigDir != "" {
			outputDir = filepath.Join(g.ctx.CoreConfig.ConfigDir, outputDir)
		}
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create protobuf output directory %s: %w", outputDir, err)
	} // Determine schema directory relative to config directory
	schemaDir := "schema"
	if g.ctx.CoreConfig.ConfigDir != "" {
		schemaDir = filepath.Join(g.ctx.CoreConfig.ConfigDir, "schema")
	}

	// Run protoc command for all proto files
	args := []string{
		"-I", schemaDir, // Add import path for schema directory
		"--go_out=" + outputDir,
		"--go-grpc_out=" + outputDir,
		"--go_opt=module=" + goPackage,
		"--go-grpc_opt=module=" + goPackage,
	}

	// Add all proto files to the command
	args = append(args, protoFiles...)

	g.ctx.Logger.Debug(fmt.Sprintf("Running protoc with args: %v", args))

	cmd := exec.Command("protoc", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("protoc failed: %w\nOutput: %s", err, string(output))
	}

	g.ctx.Logger.Info(fmt.Sprintf("Generated protobuf Go files successfully from %d proto files", len(protoFiles)))
	return nil
}

// discoverProtoFiles finds all .proto files that were generated by the protobuf plugin
func (g *StubGenerator) discoverProtoFiles() ([]string, error) {
	var protoFiles []string

	// Check for single file output first (backward compatibility)
	// Only if the output pattern doesn't contain placeholders
	if g.pluginConfig != nil && g.pluginConfig.Output != "" &&
		!strings.Contains(g.pluginConfig.Output, "{") {
		outputPath := g.pluginConfig.Output
		// Resolve relative to config directory if not absolute
		if !filepath.IsAbs(outputPath) && g.ctx.CoreConfig.ConfigDir != "" {
			outputPath = filepath.Join(g.ctx.CoreConfig.ConfigDir, outputPath)
		}
		if _, err := os.Stat(outputPath); err == nil {
			protoFiles = append(protoFiles, outputPath)
			return protoFiles, nil
		}
	}

	// Look for multiple files in schema directory
	schemaDir := "schema"
	if g.ctx.CoreConfig.ConfigDir != "" {
		schemaDir = filepath.Join(g.ctx.CoreConfig.ConfigDir, "schema")
	}

	// First check direct schema directory for backward compatibility
	if entries, err := os.ReadDir(schemaDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".proto") {
				protoFiles = append(protoFiles, filepath.Join(schemaDir, entry.Name()))
			}
		}
	}

	// If no direct files found, check schema subdirectories (like schema/proto/)
	if len(protoFiles) == 0 {
		_ = filepath.WalkDir(schemaDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Skip directories we can't read
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".proto") {
				protoFiles = append(protoFiles, path)
			}
			return nil
		})
		// if err != nil {
		// 	// If walkdir fails, that's okay, we'll fall back to other discovery methods
		// }
	}

	// If still no files found, check config directory for .proto files
	if len(protoFiles) == 0 {
		searchDir := "."
		if g.ctx.CoreConfig.ConfigDir != "" {
			searchDir = g.ctx.CoreConfig.ConfigDir
		}
		if entries, err := os.ReadDir(searchDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".proto") {
					protoFiles = append(protoFiles, filepath.Join(searchDir, entry.Name()))
				}
			}
		}
	}

	return protoFiles, nil
}

// GenerateProtobufGoFilesStandalone runs protoc on existing proto files
// This should be called AFTER all proto files have been written to disk
func GenerateProtobufGoFilesStandalone(config *StubConfig, pluginConfig *Config) error {
	return GenerateProtobufGoFilesStandaloneWithConfigDir(config, pluginConfig, "")
}

// GenerateProtobufGoFilesStandaloneWithConfigDir runs protoc on existing proto files with config directory
// This should be called AFTER all proto files have been written to disk
func GenerateProtobufGoFilesStandaloneWithConfigDir(config *StubConfig, pluginConfig *Config, configDir string) error {
	if config == nil || !config.Enabled {
		return nil
	}

	// Create a minimal context for logging with ConfigDir
	coreConfig := &parser.CoreConfig{
		ConfigDir: configDir,
	}

	ctx := &goschemagen.GenerationContext{
		GenerationContext: parser.GenerationContext{
			Logger:     parser.NewDefaultLogger(),
			CoreConfig: coreConfig,
		},
	}

	// Create a minimal stub generator just for protoc execution
	stubGen := &StubGenerator{
		config:       config,
		pluginConfig: pluginConfig,
		ctx:          ctx,
	}

	return stubGen.generateProtobufGoFiles()
}
