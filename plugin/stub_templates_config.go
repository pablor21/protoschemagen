package plugin

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplateData holds the data passed to templates
type TemplateData struct {
	ModulePath      string
	ProtobufPackage string
	ProtobufAlias   string
	PackageImports  []string // Dynamic package imports
	Types           []*TemplateTypeInfo
	Services        []*ServiceInfo
	MapConversions  []*MapConversionInfo
}

// TemplateTypeInfo represents type information for templates
type TemplateTypeInfo struct {
	Name         string
	PackageAlias string
	IsEnum       bool
	Fields       []*TemplateFieldInfo
}

// TemplateFieldInfo represents field information for templates
type TemplateFieldInfo struct {
	Name           string // Field name from gonnotation
	GoName         string // Go field name from gonnotation
	ProtoFieldName string // Protobuf Go struct field name (PascalCase)
	ProtoName      string // Protobuf schema field name (snake_case)
	Type           string
	ProtoType      string
	Tag            string // Struct tag
	IsRepeated     bool
	IsEmbedded     bool
	// Conversion functions
	ToProtoConversion   string // Function to convert from original to protobuf
	FromProtoConversion string // Function to convert from protobuf to original
}

// MapConversionInfo holds information about map conversion functions
type MapConversionInfo struct {
	ToProtoFuncName     string
	FromProtoFuncName   string
	OriginalType        string
	ProtoType           string
	ValueIsPointer      bool
	ValueConversionFunc string
	ValueFromProtoFunc  string
}

// executeTemplateByName executes a template by name using the template manager
func (g *StubGenerator) executeTemplateByName(templateName string, data *TemplateData) ([]byte, error) {
	tmpl, err := g.templateManager.GetTemplate(templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to get template %s: %w", templateName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.Bytes(), nil
}

// // executeTemplate executes a template with the given data (legacy method)
// func (g *StubGenerator) executeTemplate(templateStr string, data *TemplateData) ([]byte, error) {
// 	tmpl, err := template.New("stub").Parse(templateStr)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse template: %w", err)
// 	}

// 	var buf bytes.Buffer
// 	if err := tmpl.Execute(&buf, data); err != nil {
// 		return nil, fmt.Errorf("failed to execute template: %w", err)
// 	}

// 	return buf.Bytes(), nil
// }

// prepareTemplateData prepares data for template execution
func (g *StubGenerator) prepareTemplateData(config *TemplateConfig) *TemplateData {
	// Get unique package imports from analyzed types
	packageImports := g.extractPackageImports()

	// Auto-detect module path from go.mod or context
	modulePath := config.ModulePath
	if modulePath == "" {
		// Try to get module path from go.mod file
		if modPath := g.getModulePathFromGoMod(); modPath != "" {
			modulePath = modPath
		} else if g.ctx != nil && len(g.ctx.Structs) > 0 {
			// Fallback: extract base module path from context struct package paths
			for _, structInfo := range g.ctx.Structs {
				if structInfo.PackagePath != "" {
					// Remove the last segment (usually "models") to get base module path
					if lastSlash := strings.LastIndex(structInfo.PackagePath, "/"); lastSlash != -1 {
						modulePath = structInfo.PackagePath[:lastSlash]
						break
					}
				}
			}
		}
	}

	// Auto-detect protobuf package from generator config
	protobufPackage := config.ProtobufPackage
	if protobufPackage == "" && g.pluginConfig != nil {
		if goPackage, ok := g.pluginConfig.Options["go_package"]; ok {
			protobufPackage = goPackage
		}
	}

	// For adapters, use local import path instead of external go_package
	// This allows adapters to import from the locally generated protobuf files
	if g.config != nil && g.config.OutputDir != "" && protobufPackage != "" {
		// Extract the base path from the external go_package and replace with local path
		// e.g., "github.com/example/proto/starwars/v1" -> "github.com/example/proto/starwars/generated/proto/v1"
		if strings.Contains(protobufPackage, "/") {
			// Find the module base by removing the version suffix
			parts := strings.Split(protobufPackage, "/")
			if len(parts) > 1 {
				// Remove the last part (usually "v1") and add our output directory
				moduleParts := parts[:len(parts)-1]
				moduleBase := strings.Join(moduleParts, "/")
				protobufPackage = moduleBase + "/" + g.config.OutputDir
			}
		}
	}

	return &TemplateData{
		ModulePath:      modulePath,
		ProtobufPackage: protobufPackage,
		ProtobufAlias:   config.ProtobufAlias,
		PackageImports:  packageImports,
		Types:           g.convertToTemplateTypes(g.originalTypes),
		Services:        g.services,
		MapConversions:  g.collectMapConversions(),
	}
}

// extractPackageImports extracts unique package imports from analyzed types
func (g *StubGenerator) extractPackageImports() []string {
	packages := make(map[string]bool)

	// Add packages from original types - use PackagePath if available, otherwise Package
	for _, typeInfo := range g.originalTypes {
		if typeInfo.Package != "" && typeInfo.Package != "main" {
			// Use the full package path that's already resolved by the parser
			if packagePath := g.getPackagePath(typeInfo); packagePath != "" {
				packages[packagePath] = true
				if g.ctx != nil && g.ctx.Logger != nil {
					g.ctx.Logger.Debug(fmt.Sprintf("Added package from type %s: %s", typeInfo.Name, packagePath))
				}
			}
		}
	}

	// Add packages from services
	for _, serviceInfo := range g.services {
		if serviceInfo.Package != "" && serviceInfo.Package != "main" {
			packages[serviceInfo.Package] = true
			if g.ctx != nil && g.ctx.Logger != nil {
				g.ctx.Logger.Debug(fmt.Sprintf("Added package from service %s: %s", serviceInfo.Name, serviceInfo.Package))
			}
		}
	}

	// Filter out standard library imports - only include third-party packages
	standardLibs := map[string]bool{
		"context": true, "fmt": true, "io": true, "strings": true, "time": true,
		"errors": true, "log": true, "os": true, "path": true, "net": true,
		"http": true, "json": true, "encoding/json": true, "net/http": true,
	}

	// Convert to slice, filtering standard library packages
	result := make([]string, 0, len(packages))
	for pkg := range packages {
		// Skip standard library packages (don't contain dots or start with known stdlib prefixes)
		if standardLibs[pkg] || isStandardLibraryPackage(pkg) {
			if g.ctx != nil && g.ctx.Logger != nil {
				g.ctx.Logger.Debug(fmt.Sprintf("Filtered out standard library package: %s", pkg))
			}
			continue
		}
		if g.ctx != nil && g.ctx.Logger != nil {
			g.ctx.Logger.Debug(fmt.Sprintf("Including package import: %s", pkg))
		}
		result = append(result, pkg)
	}

	return result
}

// isStandardLibraryPackage checks if a package is from Go's standard library
func isStandardLibraryPackage(pkg string) bool {
	// Standard library packages typically don't contain dots (e.g., "fmt", "context")
	// or start with specific prefixes
	if !strings.Contains(pkg, ".") && !strings.HasPrefix(pkg, "go.") {
		return true
	}

	// Some standard library packages with prefixes
	standardPrefixes := []string{
		"archive/", "bufio/", "builtin/", "bytes/", "compress/", "container/",
		"crypto/", "database/", "debug/", "embed/", "encoding/", "errors/",
		"expvar/", "flag/", "fmt/", "go/", "hash/", "html/", "image/", "index/",
		"internal/", "io/", "log/", "math/", "mime/", "net/", "os/", "path/",
		"plugin/", "reflect/", "regexp/", "runtime/", "sort/", "strconv/",
		"strings/", "sync/", "syscall/", "testing/", "text/", "time/",
		"unicode/", "unsafe/",
	}

	for _, prefix := range standardPrefixes {
		if strings.HasPrefix(pkg, prefix) {
			return true
		}
	}

	return false
}

// getImportsForTemplate returns imports specific to a template
func (g *StubGenerator) getImportsForTemplate(templateName string, baseImports []string) []string {
	if g.ctx != nil && g.ctx.Logger != nil {
		g.ctx.Logger.Debug(fmt.Sprintf("getImportsForTemplate(%s) - base imports: %v", templateName, baseImports))
	}

	imports := make(map[string]bool)

	// Add base package imports (these are already filtered) - make a copy to avoid modifying original
	for _, imp := range baseImports {
		imports[imp] = true
	}

	// Add template-specific imports based on what's actually used
	switch templateName {
	case "bridge":
		// Only add context if there are interface services that actually need it
		needsContext := false
		for _, service := range g.services {
			if !service.IsStruct && len(service.Methods) > 0 {
				needsContext = true
				break
			}
		}
		if needsContext {
			imports["context"] = true
			if g.ctx != nil && g.ctx.Logger != nil {
				g.ctx.Logger.Debug(fmt.Sprintf("getImportsForTemplate(%s) - adding context for interface services", templateName))
			}
		}
		// Bridge doesn't use io, that's for client/adapter streaming

	case "service":
		// Service interfaces in our template don't actually use context
		// They define original Go type methods without context parameters
		// No additional imports needed for the service template

	case "adapter", "client":
		// Adapters and clients always need context
		if len(g.services) > 0 {
			imports["context"] = true
			if g.ctx != nil && g.ctx.Logger != nil {
				g.ctx.Logger.Debug(fmt.Sprintf("getImportsForTemplate(%s) - adding context for services", templateName))
			}
		}

		// Add io only for streaming methods in adapter/client
		needsIO := false
		for _, service := range g.services {
			for _, method := range service.Methods {
				if method.IsStreaming {
					needsIO = true
					break
				}
			}
			if needsIO {
				break
			}
		}
		if needsIO {
			imports["io"] = true
			if g.ctx != nil && g.ctx.Logger != nil {
				g.ctx.Logger.Debug(fmt.Sprintf("getImportsForTemplate(%s) - adding io for streaming", templateName))
			}
		}
	}

	// Convert back to slice and sort for consistent output
	result := make([]string, 0, len(imports))
	for imp := range imports {
		result = append(result, imp)
	}

	if g.ctx != nil && g.ctx.Logger != nil {
		g.ctx.Logger.Debug(fmt.Sprintf("getImportsForTemplate(%s) - final imports: %v", templateName, result))
	}

	return result
} // getPackagePath returns the correct import path for a type
func (g *StubGenerator) getPackagePath(typeInfo *TypeInfo) string {
	// Check if the type has a FullName which might contain the full path
	if typeInfo.FullName != "" {
		// Extract package path from FullName (remove the type name)
		if lastDot := strings.LastIndex(typeInfo.FullName, "."); lastDot != -1 {
			return typeInfo.FullName[:lastDot]
		}
	}

	// Fall back to Package field
	return typeInfo.Package
}

// convertToTemplateTypes converts map of TypeInfo to template-friendly format
func (g *StubGenerator) convertToTemplateTypes(types map[string]*TypeInfo) []*TemplateTypeInfo {
	result := make([]*TemplateTypeInfo, 0, len(types))
	for _, typeInfo := range types {
		templateFields := make([]*TemplateFieldInfo, len(typeInfo.Fields))
		for j, field := range typeInfo.Fields {
			// Use the ProtoName that was already determined by analyzeField
			protoName := field.ProtoName
			if protoName == "" {
				// Extract from JSON tag if available
				if jsonTag := g.extractJSONTag(field.Tag); jsonTag != "" {
					protoName = jsonTag
				} else {
					protoName = g.toSnakeCase(field.GoName) // Use GoName for conversion
				}
			}

			// Convert protobuf schema name to Go struct field name (PascalCase)
			// This should match what protoc generates
			protoFieldName := g.protoFieldNameFromTag(protoName)

			templateFields[j] = &TemplateFieldInfo{
				Name:                field.Name,
				GoName:              field.GoName,
				ProtoFieldName:      protoFieldName,
				ProtoName:           protoName,
				Type:                field.Type,
				ProtoType:           field.ProtoType,
				Tag:                 field.Tag,
				IsRepeated:          field.IsRepeated,
				IsEmbedded:          field.IsEmbedded,
				ToProtoConversion:   g.getToProtoConversion(field, field.GoName),
				FromProtoConversion: g.getFromProtoConversion(field, protoFieldName),
			}
		}

		// Extract package alias from the full package path
		packageAlias := g.getPackageAlias(typeInfo.Package)

		result = append(result, &TemplateTypeInfo{
			Name:         typeInfo.Name,
			PackageAlias: packageAlias,
			IsEnum:       typeInfo.IsEnum,
			Fields:       templateFields,
		})
	}
	return result
}

// getPackageAlias returns the alias to use for a package in templates
func (g *StubGenerator) getPackageAlias(packagePath string) string {
	if packagePath == "" {
		return "models" // fallback
	}

	// Use the last part of the path as alias
	parts := strings.Split(packagePath, "/")
	return parts[len(parts)-1]
}

// extractModulePath extracts the base module path from a full package path
func (g *StubGenerator) extractModulePath(fullPackage string) string {
	// Remove the last segment to get the module path
	// e.g., "github.com/user/repo/models" -> "github.com/user/repo"
	parts := strings.Split(fullPackage, "/")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "/")
	}
	return fullPackage
}

// resolveProtobufPackage resolves the protobuf package from go_package option
func (g *StubGenerator) resolveProtobufPackage(goPackage string) string {
	// If go_package is a full import path, use it as is
	if strings.Contains(goPackage, "/") {
		return goPackage
	}

	// If it's just a directory name, prefix with detected module path
	if g.ctx != nil {
		for _, structInfo := range g.ctx.Structs {
			if structInfo.Package != "" {
				modulePath := g.extractModulePath(structInfo.Package)
				return modulePath + "/" + goPackage
			}
		}
	}

	return goPackage
}

// toPascalCase converts snake_case to PascalCase with protobuf conventions
func (g *StubGenerator) toPascalCase(s string) string {
	// Special handling for common cases that protoc handles differently
	switch strings.ToLower(s) {
	case "id":
		return "Id" // protoc generates "Id" not "ID"
	case "adopter_id":
		return "AdopterId" // protoc generates "AdopterId" not "AdopterID"
	case "created_at":
		return "CreatedAt"
	case "updated_at":
		return "UpdatedAt"
	}

	parts := strings.Split(s, "_")
	result := ""
	for i, part := range parts {
		if len(part) > 0 {
			if i == len(parts)-1 && strings.ToLower(part) == "id" {
				// Last part is "id", use "Id" instead of "ID"
				result += "Id"
			} else {
				result += strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			}
		}
	}
	return result
}

// extractJSONTag extracts field name from json struct tag
func (g *StubGenerator) extractJSONTag(tag string) string {
	if tag == "" {
		return ""
	}

	// Look for json:"field_name" pattern
	if strings.Contains(tag, `json:"`) {
		start := strings.Index(tag, `json:"`) + 6
		end := strings.Index(tag[start:], `"`)
		if end > 0 {
			jsonTag := tag[start : start+end]
			// Remove omitempty and other options
			parts := strings.Split(jsonTag, ",")
			if len(parts) > 0 && parts[0] != "-" {
				return parts[0]
			}
		}
	}
	return ""
}

// protoFieldNameFromTag converts proto field name to Go protobuf field name using protoc conventions
func (g *StubGenerator) protoFieldNameFromTag(protoName string) string {
	if protoName == "" {
		return ""
	}

	// Handle specific protoc naming patterns
	switch protoName {
	case "id":
		return "Id"
	case "adopter_id":
		return "AdopterId"
	case "photo_urls":
		return "PhotoUrls"
	case "created_at":
		return "CreatedAt"
	case "updated_at":
		return "UpdatedAt"
	}

	// General conversion: snake_case to PascalCase with protoc rules
	parts := strings.Split(protoName, "_")
	result := ""
	for i, part := range parts {
		if len(part) > 0 {
			if i == len(parts)-1 && strings.ToLower(part) == "id" {
				result += "Id" // protoc uses "Id" not "ID" at end
			} else {
				result += strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			}
		}
	}
	return result
}

// getToProtoConversion determines the conversion function needed to convert from original to protobuf
func (g *StubGenerator) getToProtoConversion(field *FieldInfo, goFieldName string) string {
	// Check for generic type parameters and resolve them
	if resolvedType := g.resolveGenericType(field); resolvedType != field.Type {
		// Create a temporary field with resolved type for recursive processing
		resolvedField := *field
		resolvedField.Type = resolvedType
		return g.getToProtoConversion(&resolvedField, goFieldName)
	}

	// Check if this is an enum field
	if g.isEnumField(field.Type) {
		// Extract enum type name from field type
		enumName := g.extractTypeName(field.Type)

		// Check if it's a slice
		if strings.HasPrefix(field.Type, "[]") {
			return enumName + "SliceToProto(orig." + goFieldName + ")"
		} else {
			return enumName + "ToProto(orig." + goFieldName + ")"
		}
	}

	// Handle maps
	if strings.HasPrefix(field.Type, "map[") {
		return g.getMapToProtoConversion(field.Type, goFieldName)
	}

	// Handle slices of non-enum types
	if strings.HasPrefix(field.Type, "[]") {
		return g.getSliceToProtoConversion(field.Type, goFieldName)
	}

	// Handle struct types that need conversion
	if g.isStructType(field.Type) {
		typeName := g.extractTypeName(field.Type)
		if strings.HasPrefix(field.Type, "*") {
			return g.getPointerToProtoConversion(typeName, goFieldName)
		} else {
			return typeName + "ToProto(orig." + goFieldName + ")"
		}
	}

	// For simple types, direct assignment
	return "orig." + goFieldName
}

// getFromProtoConversion determines the conversion function needed to convert from protobuf to original
func (g *StubGenerator) getFromProtoConversion(field *FieldInfo, protoFieldName string) string {
	// Check for generic type parameters and resolve them
	if resolvedType := g.resolveGenericType(field); resolvedType != field.Type {
		// Create a temporary field with resolved type for recursive processing
		resolvedField := *field
		resolvedField.Type = resolvedType
		return g.getFromProtoConversion(&resolvedField, protoFieldName)
	}

	// Check if this is an enum field
	if g.isEnumField(field.Type) {
		// Extract enum type name from field type
		enumName := g.extractTypeName(field.Type)

		// Check if it's a slice
		if strings.HasPrefix(field.Type, "[]") {
			return enumName + "SliceFromProto(proto." + protoFieldName + ")"
		} else {
			return enumName + "FromProto(proto." + protoFieldName + ")"
		}
	}

	// Handle maps
	if strings.HasPrefix(field.Type, "map[") {
		return g.getMapFromProtoConversion(field.Type, protoFieldName)
	}

	// Handle slices of non-enum types
	if strings.HasPrefix(field.Type, "[]") {
		return g.getSliceFromProtoConversion(field.Type, protoFieldName)
	}

	// Handle struct types that need conversion
	if g.isStructType(field.Type) {
		typeName := g.extractTypeName(field.Type)
		if strings.HasPrefix(field.Type, "*") {
			return g.getPointerFromProtoConversion(typeName, protoFieldName)
		} else {
			return typeName + "FromProto(proto." + protoFieldName + ")"
		}
	}

	// For simple types, direct assignment
	return "proto." + protoFieldName
}

// isEnumField checks if a field type is an enum or slice of enums
func (g *StubGenerator) isEnumField(fieldType string) bool {
	// Remove slice prefix if present
	typeName := strings.TrimPrefix(fieldType, "[]")

	// Check if we have this type as an enum in our original types
	if typeInfo, exists := g.originalTypes[typeName]; exists {
		return typeInfo.IsEnum
	}

	return false
}

// extractTypeName extracts the type name from a field type (removes [] prefix, package prefix, etc.)
func (g *StubGenerator) extractTypeName(fieldType string) string {
	// Remove slice prefix if present
	typeName := strings.TrimPrefix(fieldType, "[]")

	// Remove package prefix if present (e.g., "models.Episode" -> "Episode")
	if lastDot := strings.LastIndex(typeName, "."); lastDot != -1 {
		typeName = typeName[lastDot+1:]
	}

	return typeName
}

// getMapToProtoConversion handles map type conversions
func (g *StubGenerator) getMapToProtoConversion(mapType, goFieldName string) string {
	// Extract key and value types from map[key]value
	if matches := g.parseMapType(mapType); matches != nil {
		keyType, valueType := matches[0], matches[1]

		// For simple key types and struct value types, we need conversion
		if g.isStructType(valueType) {
			// Sanitize type names for function names
			sanitizedKey := g.sanitizeTypeName(keyType)
			sanitizedValue := g.sanitizeTypeName(valueType)
			return fmt.Sprintf("ConvertMapToProto_%s_%s(orig.%s)", sanitizedKey, sanitizedValue, goFieldName)
		}
	}

	// Default to direct assignment for simple maps
	return "orig." + goFieldName
}

// getMapFromProtoConversion handles map type conversions from proto
func (g *StubGenerator) getMapFromProtoConversion(mapType, protoFieldName string) string {
	// Extract key and value types from map[key]value
	if matches := g.parseMapType(mapType); matches != nil {
		keyType, valueType := matches[0], matches[1]

		// For simple key types and struct value types, we need conversion
		if g.isStructType(valueType) {
			// Sanitize type names for function names
			sanitizedKey := g.sanitizeTypeName(keyType)
			sanitizedValue := g.sanitizeTypeName(valueType)
			return fmt.Sprintf("ConvertMapFromProto_%s_%s(proto.%s)", sanitizedKey, sanitizedValue, protoFieldName)
		}
	}

	// Default to direct assignment for simple maps
	return "proto." + protoFieldName
} // getSliceToProtoConversion handles slice type conversions
func (g *StubGenerator) getSliceToProtoConversion(sliceType, goFieldName string) string {
	elementType := strings.TrimPrefix(sliceType, "[]")

	if g.isStructType(elementType) {
		typeName := g.extractTypeName(elementType)
		return fmt.Sprintf("%sSliceToProto(orig.%s)", typeName, goFieldName)
	}

	// Default to direct assignment for simple slices
	return "orig." + goFieldName
}

// getSliceFromProtoConversion handles slice type conversions from proto
func (g *StubGenerator) getSliceFromProtoConversion(sliceType, protoFieldName string) string {
	elementType := strings.TrimPrefix(sliceType, "[]")

	if g.isStructType(elementType) {
		typeName := g.extractTypeName(elementType)
		return fmt.Sprintf("%sSliceFromProto(proto.%s)", typeName, protoFieldName)
	}

	// Default to direct assignment for simple slices
	return "proto." + protoFieldName
}

// getPointerToProtoConversion handles pointer type conversions
func (g *StubGenerator) getPointerToProtoConversion(typeName, goFieldName string) string {
	return fmt.Sprintf("ConvertPointerToProto_%s(orig.%s)", typeName, goFieldName)
}

// getPointerFromProtoConversion handles pointer type conversions from proto
func (g *StubGenerator) getPointerFromProtoConversion(typeName, protoFieldName string) string {
	return fmt.Sprintf("ConvertPointerFromProto_%s(proto.%s)", typeName, protoFieldName)
}

// isStructType checks if a type is a struct (not a primitive type)
func (g *StubGenerator) isStructType(typeName string) bool {
	// Clean up the type name
	cleanType := strings.TrimPrefix(typeName, "*")
	cleanType = strings.TrimPrefix(cleanType, "[]")

	// Check if it's a primitive type
	primitives := map[string]bool{
		"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "bool": true, "byte": true, "rune": true,
		"time.Time": true, // Common non-primitive that doesn't need conversion
	}

	if primitives[cleanType] {
		return false
	}

	// Check if it's a known type in our original types that's not an enum
	if typeInfo, exists := g.originalTypes[cleanType]; exists {
		return !typeInfo.IsEnum // Structs are non-enum types we know about
	}

	// If it contains a dot or is a capitalized identifier, it's likely a struct
	return strings.Contains(cleanType, ".") || (len(cleanType) > 0 && cleanType[0] >= 'A' && cleanType[0] <= 'Z')
} // parseMapType extracts key and value types from a map type string
func (g *StubGenerator) parseMapType(mapType string) []string {
	// Simple regex to parse map[key]value
	if strings.HasPrefix(mapType, "map[") && strings.Contains(mapType, "]") {
		content := strings.TrimPrefix(mapType, "map[")
		bracketIndex := strings.Index(content, "]")
		if bracketIndex > 0 {
			keyType := content[:bracketIndex]
			valueType := content[bracketIndex+1:]
			return []string{keyType, valueType}
		}
	}
	return nil
}

// sanitizeTypeName removes characters that can't be used in function names
func (g *StubGenerator) sanitizeTypeName(typeName string) string {
	// Replace * with Star (not Ptr to match our template)
	result := strings.ReplaceAll(typeName, "*", "Star")
	// Replace [] with Slice
	result = strings.ReplaceAll(result, "[]", "Slice")
	// Replace dots with underscores for package names
	result = strings.ReplaceAll(result, ".", "_")
	return result
}

// resolveGenericType resolves generic type parameters to their concrete types
// For example, in HumanResponse = Response[Human], field type "T" resolves to "Human"
func (g *StubGenerator) resolveGenericType(field *FieldInfo) string {
	// If it's not a generic type parameter, return as-is
	if !g.isGenericTypeParameter(field.Type) {
		return field.Type
	}

	// Look for type aliases that instantiate this generic
	// For example: HumanResponse = Response[Human]
	for typeName, typeInfo := range g.originalTypes {
		if g.isGenericTypeAlias(typeInfo, field) {
			if concreteType := g.extractConcreteTypeFromAlias(typeName, field.Type); concreteType != "" {
				return concreteType
			}
		}
	}

	// If we can't resolve it, return the original type
	return field.Type
}

// isGenericTypeParameter checks if a type is a generic parameter (single letter, typically T, K, V)
func (g *StubGenerator) isGenericTypeParameter(typeName string) bool {
	// Generic type parameters are typically single uppercase letters
	return len(typeName) == 1 && typeName >= "A" && typeName <= "Z"
}

// isGenericTypeAlias checks if a type is a generic type alias that might resolve our field
func (g *StubGenerator) isGenericTypeAlias(typeInfo *TypeInfo, field *FieldInfo) bool {
	// This is a heuristic - look for types that contain fields with the same name as our field
	for _, fieldInfo := range typeInfo.Fields {
		if fieldInfo.GoName == field.GoName && g.isGenericTypeParameter(fieldInfo.Type) {
			return true
		}
	}
	return false
}

// extractConcreteTypeFromAlias extracts the concrete type from a type alias
// For HumanResponse = Response[Human], this extracts "Human" for type parameter "T"
func (g *StubGenerator) extractConcreteTypeFromAlias(aliasName, genericParam string) string {
	// This is a simplified implementation
	// In a real-world scenario, we'd need proper AST analysis

	// Common patterns: XxxResponse = Response[Xxx]
	if strings.HasSuffix(aliasName, "Response") {
		baseType := strings.TrimSuffix(aliasName, "Response")
		if baseType != "" {
			return baseType
		}
	}

	// Pattern: XxxResult = Result[Xxx]
	if strings.HasSuffix(aliasName, "Result") {
		baseType := strings.TrimSuffix(aliasName, "Result")
		if baseType != "" {
			return baseType
		}
	}

	// Pattern: XxxEdge = Edge[Xxx]
	if strings.HasSuffix(aliasName, "Edge") {
		baseType := strings.TrimSuffix(aliasName, "Edge")
		if baseType != "" {
			return baseType
		}
	}

	return ""
}

// collectMapConversions collects all map conversion functions needed
func (g *StubGenerator) collectMapConversions() []*MapConversionInfo {
	conversions := make(map[string]*MapConversionInfo)

	// Go through all types and find map fields that need conversion
	for _, typeInfo := range g.originalTypes {
		for _, field := range typeInfo.Fields {
			if strings.HasPrefix(field.Type, "map[") {
				if mapInfo := g.parseMapType(field.Type); mapInfo != nil {
					keyType, valueType := mapInfo[0], mapInfo[1]

					// Only generate conversions for maps with struct values
					if g.isStructType(valueType) {
						sanitizedKey := g.sanitizeTypeName(keyType)
						sanitizedValue := g.sanitizeTypeName(valueType)

						functionKey := fmt.Sprintf("%s_%s", sanitizedKey, sanitizedValue)

						if _, exists := conversions[functionKey]; !exists {
							valueIsPointer := strings.HasPrefix(valueType, "*")
							cleanValueType := strings.TrimPrefix(valueType, "*")

							// Get the simple type name for protobuf and function names
							simpleTypeName := g.extractTypeName(cleanValueType)

							// For adapter generation, always use package aliases for better readability
							// Extract the simple type name and use the package alias
							simpleType := g.extractTypeName(cleanValueType)

							// Extract just the package name from the full package path
							packageAlias := typeInfo.Package
							if lastSlash := strings.LastIndex(packageAlias, "/"); lastSlash != -1 {
								packageAlias = packageAlias[lastSlash+1:]
							}

							var originalMapType string
							if valueIsPointer {
								originalMapType = fmt.Sprintf("map[%s]*%s.%s", keyType, packageAlias, simpleType)
							} else {
								originalMapType = fmt.Sprintf("map[%s]%s.%s", keyType, packageAlias, simpleType)
							}

							conversions[functionKey] = &MapConversionInfo{
								ToProtoFuncName:     fmt.Sprintf("ConvertMapToProto_%s", functionKey),
								FromProtoFuncName:   fmt.Sprintf("ConvertMapFromProto_%s", functionKey),
								OriginalType:        originalMapType,
								ProtoType:           fmt.Sprintf("map[%s]*pb.%s", keyType, simpleTypeName),
								ValueIsPointer:      valueIsPointer,
								ValueConversionFunc: fmt.Sprintf("%sToProto", simpleTypeName),
								ValueFromProtoFunc:  fmt.Sprintf("%sFromProto", simpleTypeName),
							}
						}
					}
				}
			}
		}
	}

	// Convert map to slice
	result := make([]*MapConversionInfo, 0, len(conversions))
	for _, conv := range conversions {
		result = append(result, conv)
	}

	return result
}

// getModulePathFromGoMod reads the module path from go.mod file
func (g *StubGenerator) getModulePathFromGoMod() string {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Look for go.mod file in current and parent directories
	dir := cwd
	for i := 0; i < 10; i++ { // Increase search limit
		goModPath := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(goModPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						return parts[1]
					}
				}
			}
		}
		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}
	return ""
}
