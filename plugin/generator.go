package plugin

import (
	"fmt"
	"go/ast"
	"sort"
	"strconv"
	"strings"

	"github.com/pablor21/gonnotation/annotations"
	"github.com/pablor21/gonnotation/parser"
	"github.com/pablor21/goschemagen"
)

// Generator generates Protobuf schemas
type Generator struct {
	formatGen     *Plugin
	ctx           *goschemagen.GenerationContext
	fieldNumbers  map[string]int // Track field numbers per message
	currentNumber int            // Current field number counter
	currentFile   string         // Current file being generated (for imports)
	// fieldProcessor *goschemagen.FieldProcessor
}

// getMergedKnownTypes merges core known types with format-specific known types
// Format-specific types override core types with the same schema name
func (g *Generator) getMergedKnownTypes() map[string]goschemagen.KnownTypeMapping {
	merged := make(map[string]goschemagen.KnownTypeMapping)

	// Start with core known types
	if g.formatGen.config.KnownTypes != nil {
		for name, mapping := range g.formatGen.config.KnownTypes {
			merged[name] = mapping
		}
	}

	// Protobuf doesn't have its own KnownTypes config yet
	// Just return core known types for now

	return merged
}

// validate performs validation checks before generation
func (g *Generator) Generate() ([]byte, error) {
	var out strings.Builder
	g.fieldNumbers = make(map[string]int)
	g.currentNumber = g.formatGen.config.StartFieldNumber

	// Run validation before generation
	validator := NewProtoValidator(g.ctx.Logger)
	if validationErrors := validator.ValidateContext(g.ctx, g); len(validationErrors) > 0 {
		// Check if any errors are severity "error"
		for _, ve := range validationErrors {
			if ve.Severity == "error" {
				// Log all validation errors
				for _, e := range validationErrors {
					g.ctx.Logger.Error(fmt.Sprintf("[%s] %s: %s", e.Severity, e.Location, e.Message))
				}
				return nil, fmt.Errorf("validation failed with %d error(s)", len(validationErrors))
			}
		}
	}

	// Write syntax declaration
	out.WriteString(fmt.Sprintf("syntax = \"%s\";\n\n", g.formatGen.config.Syntax))

	// Write package declaration
	if pkg := g.getPackageName(); pkg != "" {
		out.WriteString(fmt.Sprintf("package %s;\n\n", pkg))
	}

	// Write options
	if err := g.writeOptions(&out); err != nil {
		return nil, err
	}

	// Write imports
	if err := g.writeImports(&out); err != nil {
		return nil, err
	}

	// Generate messages from structs
	if err := g.generateMessages(&out); err != nil {
		return nil, err
	}

	// Generate enums
	if err := g.generateEnums(&out); err != nil {
		return nil, err
	}

	// Generate services (if enabled)
	if g.formatGen.config.GenerateService {
		if err := g.generateServices(&out); err != nil {
			return nil, err
		}
	}

	return []byte(out.String()), nil
}

func (g *Generator) getPackageName() string {
	// Fall back to config
	if g.formatGen.config.Package != "" {
		return g.formatGen.config.Package
	}
	// Default to first package name
	if g.ctx != nil && len(g.ctx.Structs) > 0 {
		if pkg := g.ctx.Structs[0].Package; pkg != "" {
			return pkg
		}
	}
	return ""
}

func (g *Generator) writeOptions(out *strings.Builder) error {
	// Collect all options (config + file-level annotations)
	options := make(map[string]string)

	// Add standard options from config (backward compatibility)
	if g.formatGen.config.GoPackage != "" {
		options["go_package"] = g.formatGen.config.GoPackage
	}
	if g.formatGen.config.JavaPackage != "" {
		options["java_package"] = g.formatGen.config.JavaPackage
	}
	if g.formatGen.config.JavaOuterClass != "" {
		options["java_outer_classname"] = g.formatGen.config.JavaOuterClass
	}
	if g.formatGen.config.OptimizeFor != "" {
		// optimize_for doesn't use quotes
		fmt.Fprintf(out, "option optimize_for = %s;\n", g.formatGen.config.OptimizeFor)
	}

	// Add additional options from config.Options
	if g.formatGen.config.Options != nil {
		for k, v := range g.formatGen.config.Options {
			options[k] = v
		}
	}

	// Extract file-level @proto.option annotations
	fileOptions := g.extractFileLevelOptions()
	for k, v := range fileOptions {
		options[k] = v // File-level overrides config
	}

	// Write all options in sorted order for consistency
	keys := make([]string, 0, len(options))
	for k := range options {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(out, "option %s = \"%s\";\n", k, options[k])
	}

	if len(options) > 0 {
		out.WriteString("\n")
	}
	return nil
}

// extractFileLevelOptions extracts @proto.option annotations from file-level comments
// Note: Currently returns empty map as file-level annotations are not yet fully implemented
// Use the config.Options map instead for now
func (g *Generator) extractFileLevelOptions() map[string]string {
	options := make(map[string]string)
	// TODO: Implement file-level @proto.option extraction when file AST is available in context
	return options
}

func (g *Generator) writeImports(out *strings.Builder) error {
	imports := make(map[string]bool)

	// Add file imports for multi-file generation
	if len(g.ctx.FileTypeMappings) > 0 {
		currentFile := g.getCurrentFileName()
		usedTypes := g.getUsedTypes()
		requiredImports := g.ctx.GetRequiredImports(currentFile, usedTypes)

		for _, importFile := range requiredImports {
			imports[importFile] = true
		}
	}

	// Add well-known types if enabled
	if g.formatGen.config.KnownTypes != nil {
		// Get merged known types from core and format configs
		knownTypes := g.getMergedKnownTypes()

		// Scan for types that match known type mappings
		for _, s := range g.ctx.Structs {
			for _, f := range s.Fields {
				goType := g.getGoTypeName(f.Type)

				// Check if this type matches any known type mapping
				for _, mapping := range knownTypes {
					for _, model := range mapping.Model {
						if strings.Contains(goType, model) {
							// Check if the mapping specifies a protobuf type that needs an import
							if mapping.Type == "google.protobuf.Timestamp" || strings.Contains(mapping.Type, "Timestamp") {
								imports["google/protobuf/timestamp.proto"] = true
							}
							if mapping.Type == "google.protobuf.Duration" || strings.Contains(mapping.Type, "Duration") {
								imports["google/protobuf/duration.proto"] = true
							}
							if mapping.Type == "google.protobuf.Empty" || strings.Contains(mapping.Type, "Empty") {
								imports["google/protobuf/empty.proto"] = true
							}
						}
					}
				}

				// Fallback to simple string matching for backward compatibility
				if strings.Contains(goType, "time.Time") {
					imports["google/protobuf/timestamp.proto"] = true
				}
				if strings.Contains(goType, "time.Duration") {
					imports["google/protobuf/duration.proto"] = true
				}

				// Check for Any type
				if strings.Contains(goType, "any") || goType == "any" {
					imports["google/protobuf/any.proto"] = true
				}

				// Check for wrapper types
				if strings.Contains(goType, "wrapperspb") {
					imports["google/protobuf/wrappers.proto"] = true
				}

				// Check for struct types
				if strings.Contains(goType, "structpb") {
					imports["google/protobuf/struct.proto"] = true
				}

				// Check for empty type
				if strings.Contains(goType, "emptypb") {
					imports["google/protobuf/empty.proto"] = true
				}
			}
		}
	}

	// Add custom imports
	for _, imp := range g.formatGen.config.CustomImports {
		imports[imp] = true
	}

	// Import annotations at package level are not available in GenerationContext.
	// Users should configure imports via generator config.

	// Write imports
	if len(imports) > 0 {
		var sortedImports []string
		for imp := range imports {
			sortedImports = append(sortedImports, imp)
		}
		sort.Strings(sortedImports)
		for _, imp := range sortedImports {
			fmt.Fprintf(out, "import \"%s\";\n", imp)
		}
		out.WriteString("\n")
	}

	return nil
}

// shouldSkipFieldForMessage determines if a field should be skipped for a specific message
// based on ignore/omit/include annotations
// Returns (shouldSkip bool, shouldReserve bool)
func (g *Generator) shouldSkipFieldForMessage(field *parser.FieldInfo, messageName string) (bool, bool) {
	// Check struct tag for "-"
	if g.ctx != nil && g.ctx.FieldProcessor != nil {
		pf := g.ctx.FieldProcessor.ProcessField(field)
		if tag := pf.Tags[parser.DerefPtr(g.ctx.CoreConfig.StructTagName, "")]; tag == "-" {
			return true, false
		}
	}

	var shouldReserve bool

	// Check annotations for ignore/omit/include
	for _, ann := range field.Annotations {
		name := strings.ToLower(ann.Name)

		// Check standalone @ignore/@skip/@omit annotations
		if name == "ignore" || name == "skip" || name == "omit" ||
			strings.HasSuffix(name, ".ignore") || strings.HasSuffix(name, ".skip") || strings.HasSuffix(name, ".omit") {
			// Check if it applies to this message
			if g.annotationAppliesToMessage(&ann, messageName, true) {
				return true, shouldReserve
			}
			continue
		}

		// Check @include annotation (inverted logic)
		if name == "include" || strings.HasSuffix(name, ".include") {
			// If include is specified, skip if it doesn't apply to this message
			if !g.annotationAppliesToMessage(&ann, messageName, false) {
				return true, shouldReserve
			}
			continue
		}

		// Check @field annotation with ignore/omit/include/reserved parameters
		if name == "field" || strings.HasSuffix(name, ".field") {
			// Check reserved parameter
			if g.hasParamForMessage(&ann, "reserved", messageName) {
				shouldReserve = true
			}

			// Check ignore parameter
			if g.hasParamForMessage(&ann, "ignore", messageName) {
				return true, shouldReserve
			}
			// Check omit parameter
			if g.hasParamForMessage(&ann, "omit", messageName) {
				return true, shouldReserve
			}
			// Check include parameter (inverted logic)
			if _, hasInclude := ann.Params["include"]; hasInclude {
				if !g.hasParamForMessage(&ann, "include", messageName) {
					return true, shouldReserve
				}
			}
		}
	}

	return false, false
}

// shouldReserveAllFieldsForMessage checks if all fields should be reserved for a specific message
// based on the reserved parameter in @proto.message annotation
func (g *Generator) shouldReserveAllFieldsForMessage(s *parser.StructInfo, messageName string) bool {
	for _, ann := range s.Annotations {
		name := strings.ToLower(ann.Name)

		// Check @proto.message annotation
		if name == "message" || strings.HasSuffix(name, ".message") {
			// Check if this annotation is for the current message
			if annName, hasName := ann.GetParamValue("name"); hasName {
				if annName != messageName {
					continue // This annotation is for a different message
				}
			}

			// Check reserved parameter as bool
			if reservedBool, ok := ann.GetParamBool("reserved"); ok {
				return reservedBool
			}

			// Check reserved parameter as string
			if reservedStr, ok := ann.GetParamValue("reserved"); ok {
				if reservedStr == "" || reservedStr == "*" || reservedStr == messageName {
					return true
				}
			}

			// Check if reserved is a string list
			if reservedList, ok := ann.GetParamStringList("reserved"); ok {
				for _, msgName := range reservedList {
					if msgName == messageName || msgName == "*" {
						return true
					}
				}
			}
		}

		// Check standalone @proto.reserved annotation
		if name == "reserved" || strings.HasSuffix(name, ".reserved") {
			// If no 'for' parameter, applies to all messages
			if _, hasFor := ann.GetParamValue("for"); !hasFor {
				return true
			}
			// Check if 'for' parameter matches this message
			if g.annotationAppliesToMessage(&ann, messageName, false) {
				return true
			}
		}
	}

	return false
} // annotationAppliesToMessage checks if an annotation applies to a specific message
// based on the 'for' parameter
func (g *Generator) annotationAppliesToMessage(ann *annotations.Annotation, messageName string, defaultAll bool) bool {
	// If no 'for' parameter, use default
	if _, hasFor := ann.Params["for"]; !hasFor {
		return defaultAll
	}

	// Check single value
	if forValue, ok := ann.GetParamValue("for"); ok {
		// Empty or "true" means all messages
		if forValue == "" || forValue == "true" || forValue == "*" {
			return true
		}
		// Check if it matches the message name
		if forValue == messageName {
			return true
		}
		// "false" means none
		if forValue == "false" {
			return false
		}
	}

	// Check array of values
	if forMessages, ok := ann.GetParamStringList("for"); ok {
		for _, msg := range forMessages {
			if msg == messageName || msg == "*" {
				return true
			}
		}
		return false
	}

	return defaultAll
}

// hasParamForMessage checks if a parameter applies to a specific message
func (g *Generator) hasParamForMessage(ann *annotations.Annotation, paramName string, messageName string) bool {
	// Check if parameter exists
	paramValue, hasParam := ann.Params[paramName]
	if !hasParam {
		return false
	}

	// Empty value or "true" means all messages
	if paramValue == "" || paramValue == "true" || paramValue == "*" {
		return true
	}

	// "false" means none
	if paramValue == "false" {
		return false
	}

	// Check if it's the message name
	if paramValue == messageName {
		return true
	}

	// Try to parse as array
	if strings.HasPrefix(paramValue, "[") && strings.HasSuffix(paramValue, "]") {
		// Remove brackets and split
		values := strings.Split(strings.Trim(paramValue, "[]"), ",")
		for _, val := range values {
			val = strings.TrimSpace(strings.Trim(val, "\"'"))
			if val == messageName || val == "*" {
				return true
			}
		}
	}

	return false
}

// fieldAnnotationAppliesToMessage checks if field annotations apply to a specific message
// based on the 'for' parameter in @proto.field annotations
func (g *Generator) fieldAnnotationAppliesToMessage(field *parser.FieldInfo, messageName string) bool {
	// Check all @proto.field annotations
	for _, ann := range field.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "field" || strings.HasSuffix(name, ".field") {
			// If 'for' parameter exists, check if it matches this message
			if _, hasFor := ann.Params["for"]; hasFor {
				// Check single value
				if forValue, ok := ann.GetParamValue("for"); ok {
					// Empty or "*" means all messages
					if forValue == "" || forValue == "*" {
						return true
					}
					// Check if it matches the message name
					if forValue == messageName {
						return true
					}
					// If it doesn't match, this annotation doesn't apply
					return false
				}

				// Check array of values
				if forMessages, ok := ann.GetParamStringList("for"); ok {
					for _, msg := range forMessages {
						if msg == messageName || msg == "*" {
							return true
						}
					}
					// Not in the list, doesn't apply
					return false
				}
			}
			// No 'for' parameter means applies to all messages
		}
	}
	return true
}

// getFieldNumberFromAnnotation extracts the field number from @proto.field annotation
func (g *Generator) getFieldNumberFromAnnotation(field *parser.FieldInfo) int {
	for _, ann := range field.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "field" || strings.HasSuffix(name, ".field") {
			if numStr := ann.Params["number"]; numStr != "" {
				if num, err := strconv.Atoi(numStr); err == nil {
					return num
				}
			}
		}
	}
	return 0
}

func (g *Generator) generateMessages(out *strings.Builder) error {
	for _, s := range g.ctx.Structs {
		if err := g.processStruct(out, s); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) processStruct(out *strings.Builder, s *parser.StructInfo) error {
	// Skip generic structs with type parameters (not concrete instantiations)
	if s.IsGeneric {
		return nil
	}

	// Check if struct should be skipped
	if g.shouldSkipType(s) {
		return nil
	}

	// Get all message names from annotations (supports multiple @proto.message)
	messageNames := g.getMessageNames(s)

	if len(messageNames) == 0 {
		// No explicit @message annotation, check if should generate default
		if !s.IsEmpty && !g.hasOtherTypeAnnotation(s) {
			// Generate with default name
			if err := g.generateMessage(out, s, ""); err != nil {
				return err
			}
		}
	} else {
		// Generate a message for each @proto.message annotation
		for _, messageName := range messageNames {
			if err := g.generateMessage(out, s, messageName); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Generator) shouldSkipType(s *parser.StructInfo) bool {
	// Check for explicit ignore/skip annotation
	for _, ann := range s.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "ignore" || name == "skip" || strings.HasSuffix(name, ".ignore") || strings.HasSuffix(name, ".skip") {
			return true
		}
	}
	return false
}

func (g *Generator) hasOtherTypeAnnotation(s *parser.StructInfo) bool {
	// Check if marked as service, enum, etc.
	for _, ann := range s.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "service" || strings.HasSuffix(name, ".service") ||
			name == "enum" || strings.HasSuffix(name, ".enum") {
			return true
		}
	}
	return false
}

// getMessageNames extracts all message names from @proto.message annotations
// Returns empty slice if no @message annotations found
func (g *Generator) getMessageNames(s *parser.StructInfo) []string {
	var names []string
	for _, ann := range s.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "message" || strings.HasSuffix(name, ".message") {
			if customName := ann.Params["name"]; customName != "" {
				names = append(names, customName)
			} else {
				// Empty name means use default
				names = append(names, "")
			}
		}
	}
	return names
}

func (g *Generator) generateMessage(out *strings.Builder, s *parser.StructInfo, messageName string) error {
	// Use provided name or get default
	if messageName == "" {
		messageName = g.getMessageName(s, "")
	}

	// Get description from the specific @proto.message annotation that matches this messageName
	desc := g.getMessageDescription(s, messageName)
	if desc != "" {
		fmt.Fprintf(out, "// %s\n", desc)
	}

	fmt.Fprintf(out, "message %s {\n", messageName)

	// Build type substitutions for generic types
	typeSubstitutions := g.buildTypeSubstitutions(s)

	// Track reserved field numbers for this message
	var reservedNumbers []int

	// Check if all fields should be reserved at message level
	messageReserveAll := g.shouldReserveAllFieldsForMessage(s, messageName)

	// Reset field number for this message
	fieldNum := g.formatGen.config.StartFieldNumber // Generate fields
	for _, f := range s.Fields {
		if f.GoName == "" || len(f.GoName) == 0 || f.GoName[0] < 'A' || f.GoName[0] > 'Z' {
			// skip unexported fields
			continue
		}

		// Check if field annotation applies to this message (via 'for' parameter)
		if !g.fieldAnnotationAppliesToMessage(f, messageName) {
			continue
		}

		// Check if field should be skipped for this specific message
		isSkipped, shouldReserve := g.shouldSkipFieldForMessage(f, messageName)

		// If message-level reserved is set and field is skipped, mark for reservation
		if isSkipped && messageReserveAll {
			shouldReserve = true
		}

		if isSkipped {
			// If reserved, track the field number
			if shouldReserve {
				if num := g.getFieldNumberFromAnnotation(f); num > 0 {
					reservedNumbers = append(reservedNumbers, num)
				}
			}
			continue
		}

		// Substitute generic types if needed
		modifiedField := f
		if len(typeSubstitutions) > 0 {
			substitutedType := g.substituteGenericType(f.Type, typeSubstitutions)
			modifiedField = &parser.FieldInfo{
				Name:        f.Name,
				GoName:      f.GoName,
				Type:        substitutedType,
				Tag:         f.Tag,
				IsEmbedded:  f.IsEmbedded,
				Annotations: f.Annotations,
			}
		}

		// Get field number from annotation or auto-assign
		num := g.getFieldNumber(modifiedField, &fieldNum)
		fieldLine := g.generateField(modifiedField, num)
		if fieldLine != "" {
			fmt.Fprintf(out, "  %s\n", fieldLine)
		}
	}

	// Output reserved field numbers if any
	if len(reservedNumbers) > 0 {
		sort.Ints(reservedNumbers)
		reservedRanges := g.compactReservedRanges(reservedNumbers)
		fmt.Fprintf(out, "  reserved %s;\n", strings.Join(reservedRanges, ", "))
	}

	// Output reserved field names from @proto.reserved annotation
	reservedNames := g.getReservedNames(s, messageName)
	if len(reservedNames) > 0 {
		quotedNames := make([]string, len(reservedNames))
		for i, name := range reservedNames {
			quotedNames[i] = fmt.Sprintf("\"%s\"", name)
		}
		fmt.Fprintf(out, "  reserved %s;\n", strings.Join(quotedNames, ", "))
	}

	out.WriteString("}\n\n")
	return nil
}

// compactReservedRanges converts a sorted list of numbers into compact ranges
// e.g., [1,2,3,5,7,8,9] -> ["1 to 3", "5", "7 to 9"]
func (g *Generator) compactReservedRanges(numbers []int) []string {
	if len(numbers) == 0 {
		return nil
	}

	var ranges []string
	start := numbers[0]
	end := numbers[0]

	for i := 1; i < len(numbers); i++ {
		if numbers[i] == end+1 {
			end = numbers[i]
		} else {
			if start == end {
				ranges = append(ranges, fmt.Sprintf("%d", start))
			} else if end == start+1 {
				ranges = append(ranges, fmt.Sprintf("%d, %d", start, end))
			} else {
				ranges = append(ranges, fmt.Sprintf("%d to %d", start, end))
			}
			start = numbers[i]
			end = numbers[i]
		}
	}

	// Add the last range
	if start == end {
		ranges = append(ranges, fmt.Sprintf("%d", start))
	} else if end == start+1 {
		ranges = append(ranges, fmt.Sprintf("%d, %d", start, end))
	} else {
		ranges = append(ranges, fmt.Sprintf("%d to %d", start, end))
	}

	return ranges
}

// getReservedNames extracts reserved field names from @proto.reserved annotation
func (g *Generator) getReservedNames(s *parser.StructInfo, messageName string) []string {
	var names []string

	for _, ann := range s.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "reserved" || strings.HasSuffix(name, ".reserved") {
			// Check if this reserved annotation applies to this message
			if forValue, hasFor := ann.GetParamValue("for"); hasFor {
				if forValue != "" && forValue != "*" && forValue != messageName {
					continue
				}
			}

			// Get reserved names
			if namesParam, ok := ann.GetParamValue("names"); ok && namesParam != "" {
				names = append(names, namesParam)
			}

			if namesList, ok := ann.GetParamStringList("names"); ok {
				names = append(names, namesList...)
			}
		}
	}

	return names
}

func (g *Generator) getMessageName(s *parser.StructInfo, overrideName string) string {
	// Use override if provided
	if overrideName != "" {
		return overrideName
	}

	// Check for name override in @message annotation (first one only for default)
	for _, ann := range s.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "message" || strings.HasSuffix(name, ".message") {
			if customName := ann.Params["name"]; customName != "" {
				return customName
			}
			break // Only check first @message annotation for default
		}
	}
	return s.Name
}

// buildTypeSubstitutions creates a map of generic type parameters to concrete types
// For alias instantiations like: type HumanResponse = Response[Human]
func (g *Generator) buildTypeSubstitutions(s *parser.StructInfo) map[string]ast.Expr {
	typeSubstitutions := make(map[string]ast.Expr)

	// Only process alias instantiations
	if !s.IsAliasInstantiation || len(s.AliasTypeArgs) == 0 {
		return typeSubstitutions
	}

	// Find the generic type definition
	genericStruct := g.findStructInAST(s.AliasTarget)
	if genericStruct == nil {
		return typeSubstitutions
	}

	if genericStruct.TypeSpec == nil {
		return typeSubstitutions
	}

	if genericStruct.TypeSpec.TypeParams == nil {
		return typeSubstitutions
	} // Parse type args from strings to AST expressions
	var typeArgExprs []ast.Expr
	for _, argStr := range s.AliasTypeArgs {
		if strings.HasPrefix(argStr, "[]") {
			// Array type
			elementType := strings.TrimPrefix(argStr, "[]")
			typeArgExprs = append(typeArgExprs, &ast.ArrayType{
				Elt: &ast.Ident{Name: elementType},
			})
		} else {
			// Simple type
			typeArgExprs = append(typeArgExprs, &ast.Ident{Name: argStr})
		}
	}

	// Map type parameters to concrete types
	params := genericStruct.TypeSpec.TypeParams.List
	for i, param := range params {
		if i < len(typeArgExprs) {
			for _, name := range param.Names {
				typeSubstitutions[name.Name] = typeArgExprs[i]
			}
		}
	}

	return typeSubstitutions
}

// findStructInAST finds a struct definition by name in the context
// Searches both Structs and AllStructs to find generic definitions
func (g *Generator) findStructInAST(name string) *parser.StructInfo {
	// First check regular structs
	for _, s := range g.ctx.Structs {
		if s.Name == name {
			return s
		}
	}
	// Also check AllStructs (includes generic structs that were filtered out)
	if g.ctx.AllStructs != nil {
		for _, s := range g.ctx.AllStructs {
			if s.Name == name {
				return s
			}
		}
	}
	return nil
}

// substituteGenericType replaces type parameters with concrete types
func (g *Generator) substituteGenericType(fieldType ast.Expr, substitutions map[string]ast.Expr) ast.Expr {
	switch t := fieldType.(type) {
	case *ast.Ident:
		// Check if this is a type parameter
		if sub, ok := substitutions[t.Name]; ok {
			return sub
		}
		return fieldType

	case *ast.StarExpr:
		// Pointer type - substitute the underlying type
		return &ast.StarExpr{
			X: g.substituteGenericType(t.X, substitutions),
		}

	case *ast.ArrayType:
		// Array type - substitute the element type
		return &ast.ArrayType{
			Elt: g.substituteGenericType(t.Elt, substitutions),
		}

	case *ast.IndexExpr:
		// Generic instantiation - substitute type arguments
		return &ast.IndexExpr{
			X:     t.X,
			Index: g.substituteGenericType(t.Index, substitutions),
		}

	case *ast.IndexListExpr:
		// Multiple type arguments
		newIndices := make([]ast.Expr, len(t.Indices))
		for i, idx := range t.Indices {
			newIndices[i] = g.substituteGenericType(idx, substitutions)
		}
		return &ast.IndexListExpr{
			X:       t.X,
			Indices: newIndices,
		}

	default:
		return fieldType
	}
}

// getMessageDescription gets the description for a specific message name
// If messageName matches a @proto.message annotation's name param, use that annotation's description
// Otherwise, fall back to the general struct comment or first @message description
func (g *Generator) getMessageDescription(s *parser.StructInfo, messageName string) string {
	// First, try to find a @proto.message annotation that matches this messageName
	for _, ann := range s.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "message" || strings.HasSuffix(name, ".message") {
			annName := ann.Params["name"]
			// Match either by explicit name or by default name (struct name)
			if annName == messageName || (annName == "" && s.Name == messageName) {
				if desc := ann.Params["description"]; desc != "" {
					return desc
				}
			}
		}
	}

	// Fall back to general description from all annotations
	return g.getDescription(s.Annotations)
}

func (g *Generator) getFieldNumber(f *parser.FieldInfo, counter *int) int {
	// Check for explicit number in annotation or struct tag
	for _, ann := range f.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "field" || strings.HasSuffix(name, ".field") || name == "map" || strings.HasSuffix(name, ".map") {
			if numStr := ann.Params["number"]; numStr != "" {
				var num int
				if _, err := fmt.Sscanf(numStr, "%d", &num); err == nil && num > 0 {
					return num
				}
			}
		}
	}

	// Check struct tag via FieldProcessor tags (using configured StructTagName)
	if g.ctx != nil && g.ctx.FieldProcessor != nil {
		pf := g.ctx.FieldProcessor.ProcessField(f)
		// allow proto:number=NN in custom struct tag name
		if tag := pf.Tags[parser.DerefPtr(g.ctx.CoreConfig.StructTagName, "")]; tag != "" {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				if strings.HasPrefix(part, "number=") {
					var num int
					if _, err := fmt.Sscanf(strings.TrimPrefix(part, "number="), "%d", &num); err == nil && num > 0 {
						return num
					}
				}
			}
		}
	}

	// Auto-assign
	num := *counter
	*counter++
	return num
}

func (g *Generator) generateField(f *parser.FieldInfo, number int) string {
	fieldName := g.getFieldName(f)

	// Check if this is a map field
	if mapDef := g.generateMapField(f, fieldName, number); mapDef != "" {
		return mapDef
	}

	protoType := g.getProtoType(f)

	// Check for repeated (array/slice)
	isRepeated := g.isRepeated(f)

	// Check for optional (pointer in proto3)
	isOptional := g.isOptional(f)

	var parts []string

	// Add repeated keyword
	if isRepeated {
		parts = append(parts, "repeated")
	} else if isOptional && g.formatGen.config.Syntax == "proto3" {
		parts = append(parts, "optional")
	}

	// Add type and name
	parts = append(parts, protoType, fieldName)

	// Add field number
	fieldDef := fmt.Sprintf("%s = %d", strings.Join(parts, " "), number)

	// Add options (packed, deprecated, json_name, etc.)
	options := g.getFieldOptions(f)
	if len(options) > 0 {
		fieldDef += fmt.Sprintf(" [%s]", strings.Join(options, ", "))
	}

	fieldDef += ";"

	// Add comment
	if desc := g.getDescription(f.Annotations); desc != "" {
		fieldDef = fmt.Sprintf("// %s\n  %s", desc, fieldDef)
	}

	return fieldDef
}

// generateMapField generates a map field definition if the field is a map
func (g *Generator) generateMapField(f *parser.FieldInfo, fieldName string, number int) string {
	// Check for @proto.map annotation
	for _, ann := range f.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "map" || strings.HasSuffix(name, ".map") {
			keyType, _ := ann.GetParamValue("key")
			valueType, _ := ann.GetParamValue("value")
			if keyType != "" && valueType != "" {
				// Check for custom number in @proto.map annotation
				if numStr := ann.Params["number"]; numStr != "" {
					_, _ = fmt.Sscanf(numStr, "%d", &number) // Ignore error, keep default if parse fails
				}

				mapDef := fmt.Sprintf("map<%s, %s> %s = %d", keyType, valueType, fieldName, number)

				// Add options if any
				options := g.getFieldOptions(f)
				if len(options) > 0 {
					mapDef += fmt.Sprintf(" [%s]", strings.Join(options, ", "))
				}
				mapDef += ";"

				// Add comment
				if desc := g.getDescription(f.Annotations); desc != "" {
					mapDef = fmt.Sprintf("// %s\n  %s", desc, mapDef)
				}
				return mapDef
			}
		}
	}

	// Auto-detect Go map types
	if mapType, ok := f.Type.(*ast.MapType); ok {
		keyType := g.mapGoTypeToProto(mapType.Key)
		valueType := g.mapGoTypeToProto(mapType.Value)

		// Validate key type (protobuf only allows specific types as map keys)
		if g.isValidMapKey(keyType) {
			mapDef := fmt.Sprintf("map<%s, %s> %s = %d", keyType, valueType, fieldName, number)

			// Add options if any
			options := g.getFieldOptions(f)
			if len(options) > 0 {
				mapDef += fmt.Sprintf(" [%s]", strings.Join(options, ", "))
			}
			mapDef += ";"

			// Add comment
			if desc := g.getDescription(f.Annotations); desc != "" {
				mapDef = fmt.Sprintf("// %s\n  %s", desc, mapDef)
			}
			return mapDef
		}
	}

	return "" // Not a map field
}

// isValidMapKey checks if a type is valid as a protobuf map key
func (g *Generator) isValidMapKey(protoType string) bool {
	// Protobuf allows: int32, int64, uint32, uint64, sint32, sint64, fixed32, fixed64, sfixed32, sfixed64, bool, string
	validKeys := map[string]bool{
		"int32": true, "int64": true, "uint32": true, "uint64": true,
		"sint32": true, "sint64": true, "fixed32": true, "fixed64": true,
		"sfixed32": true, "sfixed64": true, "bool": true, "string": true,
	}
	return validKeys[protoType]
}

func (g *Generator) getFieldName(f *parser.FieldInfo) string {
	// Check annotation
	for _, ann := range f.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "field" || strings.HasSuffix(name, ".field") {
			if customName := ann.Params["name"]; customName != "" {
				return customName
			}
		}
	}

	// Check struct tag via FieldProcessor (respect CoreConfig)
	if g.ctx != nil && g.ctx.FieldProcessor != nil {
		pf := g.ctx.FieldProcessor.ProcessField(f)
		// prefer explicit proto name inside struct tag value (first segment)
		if tag := pf.Tags[parser.DerefPtr(g.ctx.CoreConfig.StructTagName, "")]; tag != "" {
			parts := strings.Split(tag, ",")
			if len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
				return parts[0]
			}
		}
	}

	// Convert Go field name using configured naming strategy
	if g.ctx != nil && g.ctx.NamingStrategy != nil {
		return goschemagen.TransformFieldName(f.Name, goschemagen.FieldCaseSnake)
	}
	return goschemagen.TransformFieldName(f.Name, goschemagen.FieldCaseSnake)
}

func (g *Generator) getProtoType(f *parser.FieldInfo) string {
	// Check for explicit type override
	for _, ann := range f.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "field" || strings.HasSuffix(name, ".field") {
			if customType := ann.Params["type"]; customType != "" {
				return customType
			}
		}
	}

	// Check struct tag via FieldProcessor
	if g.ctx != nil && g.ctx.FieldProcessor != nil {
		pf := g.ctx.FieldProcessor.ProcessField(f)
		if tag := pf.Tags[parser.DerefPtr(g.ctx.CoreConfig.StructTagName, "")]; tag != "" {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				if strings.HasPrefix(part, "type=") {
					return strings.TrimPrefix(part, "type=")
				}
			}
		}
	}

	// Map Go type to protobuf type
	return g.mapGoTypeToProto(f.Type)
}

func (g *Generator) mapGoTypeToProto(t ast.Expr) string {
	// Handle composite types structurally to avoid recursion issues
	switch v := t.(type) {
	case *ast.ArrayType:
		// Repeated is handled separately; map the element type
		return g.mapGoTypeToProto(v.Elt)
	case *ast.StarExpr:
		// Optional is handled separately; map the pointed type
		return g.mapGoTypeToProto(v.X)
	}

	goType := g.getGoTypeName(t)

	// Check custom mappings
	if mapped, ok := g.formatGen.config.TypeMappings[goType]; ok {
		return mapped
	}

	// Standard mappings
	switch goType {
	case "string", "*string":
		return "string"
	case "int", "int32", "*int32":
		return "int32"
	case "int64", "*int64":
		return "int64"
	case "uint", "uint32", "*uint32":
		return "uint32"
	case "uint64", "*uint64":
		return "uint64"
	case "bool", "*bool":
		return "bool"
	case "float32", "*float32":
		return "float"
	case "float64", "*float64":
		return "double"
	case "[]byte":
		return "bytes"
	case "time.Time", "*time.Time":
		return "google.protobuf.Timestamp"
	case "time.Duration", "*time.Duration":
		return "google.protobuf.Duration"
	case "any", "*any":
		return "google.protobuf.Any"
	case "interface{}", "*interface{}":
		return "google.protobuf.Any"

	// Well-known wrapper types
	case "wrapperspb.StringValue", "*wrapperspb.StringValue":
		return "google.protobuf.StringValue"
	case "wrapperspb.Int32Value", "*wrapperspb.Int32Value":
		return "google.protobuf.Int32Value"
	case "wrapperspb.Int64Value", "*wrapperspb.Int64Value":
		return "google.protobuf.Int64Value"
	case "wrapperspb.UInt32Value", "*wrapperspb.UInt32Value":
		return "google.protobuf.UInt32Value"
	case "wrapperspb.UInt64Value", "*wrapperspb.UInt64Value":
		return "google.protobuf.UInt64Value"
	case "wrapperspb.FloatValue", "*wrapperspb.FloatValue":
		return "google.protobuf.FloatValue"
	case "wrapperspb.DoubleValue", "*wrapperspb.DoubleValue":
		return "google.protobuf.DoubleValue"
	case "wrapperspb.BoolValue", "*wrapperspb.BoolValue":
		return "google.protobuf.BoolValue"
	case "wrapperspb.BytesValue", "*wrapperspb.BytesValue":
		return "google.protobuf.BytesValue"
	case "structpb.Struct", "*structpb.Struct":
		return "google.protobuf.Struct"
	case "structpb.Value", "*structpb.Value":
		return "google.protobuf.Value"
	case "structpb.ListValue", "*structpb.ListValue":
		return "google.protobuf.ListValue"
	case "emptypb.Empty", "*emptypb.Empty":
		return "google.protobuf.Empty"
	}

	// Basic string prefixes already handled above for structured types

	// Assume it's a message type
	return goType
}

func (g *Generator) getGoTypeName(t ast.Expr) string {
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

func (g *Generator) isRepeated(f *parser.FieldInfo) bool {
	// Check annotation
	for _, ann := range f.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "field" || strings.HasSuffix(name, ".field") {
			if repeated := ann.Params["repeated"]; repeated == "true" {
				return true
			}
		}
	}

	// Check if it's a slice/array
	goType := g.getGoTypeName(f.Type)
	return strings.HasPrefix(goType, "[]") && !strings.HasPrefix(goType, "[]byte")
}

func (g *Generator) isOptional(f *parser.FieldInfo) bool {
	// Check annotation
	for _, ann := range f.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "field" || strings.HasSuffix(name, ".field") {
			if optional := ann.Params["optional"]; optional == "true" {
				return true
			}
		}
	}

	// Check if it's a pointer
	goType := g.getGoTypeName(f.Type)
	return strings.HasPrefix(goType, "*")
}

func (g *Generator) getFieldOptions(f *parser.FieldInfo) []string {
	var options []string

	// Check for options via annotations
	for _, ann := range f.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "field" || strings.HasSuffix(name, ".field") {
			// packed option
			if packedBool, ok := ann.GetParamBool("packed"); ok && packedBool {
				options = append(options, "packed = true")
			}
			// deprecated option
			if deprecatedBool, ok := ann.GetParamBool("deprecated"); ok && deprecatedBool {
				options = append(options, "deprecated = true")
			}
			// json_name option
			if jsonName, ok := ann.GetParamValue("json_name"); ok && jsonName != "" {
				options = append(options, fmt.Sprintf("json_name = \"%s\"", jsonName))
			}
		}
		// Check @proto.option annotation for custom options
		if name == "option" || strings.HasSuffix(name, ".option") {
			if optName, ok := ann.GetParamValue("name"); ok {
				if optValue, ok := ann.GetParamValue("value"); ok {
					// Try to parse as bool or number, otherwise treat as string
					if optValue == "true" || optValue == "false" {
						options = append(options, fmt.Sprintf("%s = %s", optName, optValue))
					} else if _, err := strconv.Atoi(optValue); err == nil {
						options = append(options, fmt.Sprintf("%s = %s", optName, optValue))
					} else {
						options = append(options, fmt.Sprintf("%s = \"%s\"", optName, optValue))
					}
				}
			}
		}
	}

	// Also check struct tag via FieldProcessor
	if g.ctx != nil && g.ctx.FieldProcessor != nil {
		pf := g.ctx.FieldProcessor.ProcessField(f)
		if tag := pf.Tags[parser.DerefPtr(g.ctx.CoreConfig.StructTagName, "")]; tag != "" {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				if strings.HasPrefix(part, "json_name=") {
					jsonName := strings.TrimPrefix(part, "json_name=")
					// Check if not already added
					optStr := fmt.Sprintf("json_name = \"%s\"", jsonName)
					if !g.containsOption(options, "json_name") {
						options = append(options, optStr)
					}
				} else if strings.HasPrefix(part, "packed=") {
					if strings.TrimPrefix(part, "packed=") == "true" && !g.containsOption(options, "packed") {
						options = append(options, "packed = true")
					}
				} else if strings.HasPrefix(part, "deprecated=") {
					if strings.TrimPrefix(part, "deprecated=") == "true" && !g.containsOption(options, "deprecated") {
						options = append(options, "deprecated = true")
					}
				}
			}
		}
	}
	return options
}

// containsOption checks if an option with the given name prefix already exists
func (g *Generator) containsOption(options []string, optionName string) bool {
	prefix := optionName + " ="
	for _, opt := range options {
		if strings.HasPrefix(opt, prefix) {
			return true
		}
	}
	return false
}

func (g *Generator) getDescription(anns []annotations.Annotation) string {
	for _, ann := range anns {
		if desc := ann.Params["description"]; desc != "" {
			return desc
		}
	}
	return ""
}

// func (g *Generator) toSnakeCase(s string) string {
// 	// Smart snake_case that avoids splitting consecutive acronyms letter-by-letter
// 	if s == "" {
// 		return s
// 	}
// 	var out strings.Builder
// 	runes := []rune(s)
// 	for i, r := range runes {
// 		isUpper := r >= 'A' && r <= 'Z'
// 		var prev rune
// 		var next rune
// 		if i > 0 {
// 			prev = runes[i-1]
// 		}
// 		if i+1 < len(runes) {
// 			next = runes[i+1]
// 		}
// 		prevUpper := prev >= 'A' && prev <= 'Z'
// 		prevLower := prev >= 'a' && prev <= 'z'
// 		nextLower := next >= 'a' && next <= 'z'

// 		// Insert underscore on transitions:
// 		// - lower/digit to upper
// 		// - acronym boundary: upper followed by upper then next lower
// 		if i > 0 && isUpper && (prevLower || (prevUpper && nextLower)) {
// 			out.WriteRune('_')
// 		}
// 		out.WriteRune(r)
// 	}
// 	return strings.ToLower(out.String())
// }

func (g *Generator) generateEnums(out *strings.Builder) error {
	for _, e := range g.ctx.Enums {
		// Skip if marked as ignored
		shouldSkip := false
		for _, ann := range e.Annotations {
			name := strings.ToLower(ann.Name)
			if name == "ignore" || name == "skip" || strings.HasSuffix(name, ".ignore") || strings.HasSuffix(name, ".skip") {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}

		if err := g.generateEnum(out, e); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateEnum(out *strings.Builder, e *parser.EnumInfo) error {
	enumName := g.getEnumName(e)

	// Write comment
	if desc := g.getDescription(e.Annotations); desc != "" {
		fmt.Fprintf(out, "// %s\n", desc)
	}

	fmt.Fprintf(out, "enum %s {\n", enumName)

	// Generate enum values
	for i, v := range e.Values {
		valueNum := i // Default to index

		// Check for custom number in annotation
		for _, ann := range v.Annotations {
			name := strings.ToLower(ann.Name)
			if name == "enumvalue" || strings.HasSuffix(name, ".enumvalue") {
				if numStr := ann.Params["number"]; numStr != "" {
					_, _ = fmt.Sscanf(numStr, "%d", &valueNum) // Ignore error, keep default if parse fails
				}
			}
		}

		valueName := g.getEnumValueName(v, enumName)
		valueLine := fmt.Sprintf("  %s = %d;", valueName, valueNum)

		// Add comment
		if desc := g.getDescription(v.Annotations); desc != "" {
			valueLine = fmt.Sprintf("  // %s\n%s", desc, valueLine)
		}

		out.WriteString(valueLine + "\n")
	}

	out.WriteString("}\n\n")
	return nil
}

func (g *Generator) getEnumName(e *parser.EnumInfo) string {
	for _, ann := range e.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "enum" || strings.HasSuffix(name, ".enum") {
			if customName := ann.Params["name"]; customName != "" {
				return customName
			}
		}
	}
	return e.Name
}

func (g *Generator) getEnumValueName(v *parser.EnumValue, _ string) string {
	for _, ann := range v.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "enumvalue" || strings.HasSuffix(name, ".enumvalue") {
			if customName := ann.Params["name"]; customName != "" {
				return customName
			}
		}
	}
	// Protobuf convention: SCREAMING_SNAKE value names
	return goschemagen.TransformFieldName(v.Name, goschemagen.FieldCaseScreamingSnake)
}

func (g *Generator) generateServices(out *strings.Builder) error {
	// Generate services from interfaces with @service annotation
	for _, iface := range g.ctx.Interfaces {
		if !g.shouldGenerateService(iface) {
			continue
		}

		if err := g.generateService(out, iface); err != nil {
			return err
		}
	}

	// Generate services from structs with @service annotation
	for _, structInfo := range g.ctx.Structs {
		if !g.shouldGenerateServiceFromStruct(structInfo) {
			continue
		}

		if err := g.generateServiceFromStruct(out, structInfo); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) shouldGenerateService(iface *parser.InterfaceInfo) bool {
	for _, ann := range iface.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "service" || strings.HasSuffix(name, ".service") {
			return true
		}
	}
	return false
}

func (g *Generator) generateService(out *strings.Builder, iface *parser.InterfaceInfo) error {
	serviceName := g.getServiceName(iface)

	// Write comment
	if desc := g.getDescription(iface.Annotations); desc != "" {
		fmt.Fprintf(out, "// %s\n", desc)
	}

	fmt.Fprintf(out, "service %s {\n", serviceName)

	// Generate RPC methods from interface methods
	for _, method := range iface.Methods {
		rpcLine := g.generateRPC(method)
		if rpcLine != "" {
			fmt.Fprintf(out, "  %s\n", rpcLine)
		}
	}

	out.WriteString("}\n\n")
	return nil
}

func (g *Generator) getServiceName(iface *parser.InterfaceInfo) string {
	for _, ann := range iface.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "service" || strings.HasSuffix(name, ".service") {
			if customName := ann.Params["name"]; customName != "" {
				return customName
			}
		}
	}
	return iface.Name
}

func (g *Generator) generateRPC(method *parser.MethodInfo) string {
	// Get RPC name
	rpcName := method.Name

	// Get input/output types from function signature
	inputType := "google.protobuf.Empty"
	outputType := "google.protobuf.Empty"

	if len(method.Params) > 0 {
		inputType = g.getGoTypeName(method.Params[0].Type)
	}

	if len(method.Results) > 0 {
		outputType = g.getGoTypeName(method.Results[0].Type)
	}

	// Check for @rpc annotation overrides
	for _, ann := range method.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "rpc" || strings.HasSuffix(name, ".rpc") {
			if custom := ann.Params["name"]; custom != "" {
				rpcName = custom
			}
			if custom := ann.Params["input"]; custom != "" {
				inputType = custom
			}
			if custom := ann.Params["output"]; custom != "" {
				outputType = custom
			}
		}
	}

	// Handle streaming
	inputStream := ""
	outputStream := ""

	for _, ann := range method.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "rpc" || strings.HasSuffix(name, ".rpc") {
			if ann.Params["client_streaming"] == "true" {
				inputStream = "stream "
			}
			if ann.Params["server_streaming"] == "true" {
				outputStream = "stream "
			}
		}
	}

	return fmt.Sprintf("rpc %s(%s%s) returns (%s%s);", rpcName, inputStream, inputType, outputStream, outputType)
}

// shouldGenerateServiceFromStruct checks if a struct should be treated as a service
func (g *Generator) shouldGenerateServiceFromStruct(structInfo *parser.StructInfo) bool {
	for _, ann := range structInfo.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "service" || strings.HasSuffix(name, ".service") {
			return true
		}
	}
	return false
}

// generateServiceFromStruct generates a protobuf service from a struct with annotated methods
func (g *Generator) generateServiceFromStruct(out *strings.Builder, structInfo *parser.StructInfo) error {
	serviceName := g.getServiceNameFromStruct(structInfo)

	fmt.Fprintf(out, "\nservice %s {\n", serviceName)

	methodCount := 0
	// Look for annotated methods in the Functions list
	for _, fn := range g.ctx.Functions {
		// Check if this function is a method of our struct
		if !g.isStructMethod(fn, structInfo) {
			continue
		}

		// Check if this method has @rpc annotation
		if !g.isRPCMethod(fn) {
			continue
		}

		rpcDef, err := g.generateRPCFromFunction(fn)
		if err != nil {
			return fmt.Errorf("failed to generate RPC for method %s: %w", fn.Name, err)
		}

		fmt.Fprintf(out, "  %s\n", rpcDef)
		methodCount++
	}

	out.WriteString("}\n")
	return nil
} // getServiceNameFromStruct extracts the service name from struct annotations
func (g *Generator) getServiceNameFromStruct(structInfo *parser.StructInfo) string {
	for _, ann := range structInfo.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "service" || strings.HasSuffix(name, ".service") {
			if serviceName := ann.Params["name"]; serviceName != "" {
				return serviceName
			}
		}
	}
	return structInfo.Name
}

// isStructMethod checks if a function is a method of the given struct
func (g *Generator) isStructMethod(fn *parser.FunctionInfo, structInfo *parser.StructInfo) bool {
	// Check if function has a receiver that matches our struct
	if fn.Receiver == nil {
		return false
	}

	// Get the receiver type name
	receiverType := fn.Receiver.TypeName

	return receiverType == structInfo.Name
}

// isRPCMethod checks if a function has @rpc annotation
func (g *Generator) isRPCMethod(fn *parser.FunctionInfo) bool {
	for _, ann := range fn.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "rpc" || strings.HasSuffix(name, ".rpc") {
			return true
		}
	}
	return false
}

// generateRPCFromFunction generates an RPC definition from a function
func (g *Generator) generateRPCFromFunction(fn *parser.FunctionInfo) (string, error) {
	rpcName := fn.Name
	inputType := "google.protobuf.Empty"
	outputType := "google.protobuf.Empty"

	// Extract RPC details from annotations
	for _, ann := range fn.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "rpc" || strings.HasSuffix(name, ".rpc") {
			if custom := ann.Params["name"]; custom != "" {
				rpcName = custom
			}
			if custom := ann.Params["input"]; custom != "" {
				inputType = custom
			}
			if custom := ann.Params["output"]; custom != "" {
				outputType = custom
			}
		}
	}

	// Handle streaming
	inputStream := ""
	outputStream := ""

	for _, ann := range fn.Annotations {
		name := strings.ToLower(ann.Name)
		if name == "rpc" || strings.HasSuffix(name, ".rpc") {
			if ann.Params["client_streaming"] == "true" {
				inputStream = "stream "
			}
			if ann.Params["server_streaming"] == "true" {
				outputStream = "stream "
			}
		}
	}

	return fmt.Sprintf("rpc %s(%s%s) returns (%s%s);", rpcName, inputStream, inputType, outputStream, outputType), nil
}

// getCurrentFileName gets the current file being generated (used for imports)
func (g *Generator) getCurrentFileName() string {
	// This will be set during file generation context
	// For now, we'll need to track this in the generation context
	if g.currentFile != "" {
		return g.currentFile
	}
	return ""
}

// getUsedTypes analyzes the current structs/interfaces to find referenced types
func (g *Generator) getUsedTypes() []string {
	var usedTypes []string
	typeSet := make(map[string]bool)

	// Analyze struct fields for type references
	for _, s := range g.ctx.Structs {
		for _, f := range s.Fields {
			typeName := g.extractTypeName(g.getGoTypeName(f.Type))
			if typeName != "" && !typeSet[typeName] {
				usedTypes = append(usedTypes, typeName)
				typeSet[typeName] = true
			}
		}
	}

	// Analyze service methods for type references
	for _, i := range g.ctx.Interfaces {
		for _, m := range i.Methods {
			// Check input parameters
			for _, p := range m.Params {
				typeName := g.extractTypeName(g.getGoTypeName(p.Type))
				if typeName != "" && !typeSet[typeName] {
					usedTypes = append(usedTypes, typeName)
					typeSet[typeName] = true
				}
			}
			// Check return types
			for _, r := range m.Results {
				typeName := g.extractTypeName(g.getGoTypeName(r.Type))
				if typeName != "" && !typeSet[typeName] {
					usedTypes = append(usedTypes, typeName)
					typeSet[typeName] = true
				}
			}
		}
	}

	return usedTypes
} // extractTypeName extracts the base type name from a Go type string
func (g *Generator) extractTypeName(typeStr string) string {
	// Remove package paths and get just the type name
	if idx := strings.LastIndex(typeStr, "."); idx != -1 {
		return typeStr[idx+1:]
	}

	// Handle pointer, slice, map prefixes
	typeStr = strings.TrimPrefix(typeStr, "*")
	typeStr = strings.TrimPrefix(typeStr, "[]")
	if strings.HasPrefix(typeStr, "map[") {
		// Extract value type from map[K]V
		if idx := strings.LastIndex(typeStr, "]"); idx != -1 && idx < len(typeStr)-1 {
			typeStr = typeStr[idx+1:]
		}
	}

	// Remove package path again after prefix removal
	if idx := strings.LastIndex(typeStr, "."); idx != -1 {
		return typeStr[idx+1:]
	}

	return typeStr
}
