package plugin

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"strings"

	"github.com/pablor21/gonnotation/parser"
)

// MultiFormatGenerator handles generation of multiple output formats
type MultiFormatGenerator struct {
	generator *Generator
}

// GenerateFormat generates output in the specified format
func (mfg *MultiFormatGenerator) GenerateFormat(format string) ([]byte, string, error) {
	switch strings.ToLower(format) {
	case "json-schema", "json_schema", "jsonschema":
		content, err := mfg.generateJSONSchema()
		return content, ".schema.json", err
	case "markdown", "md":
		content, err := mfg.generateMarkdown()
		return content, ".md", err
	case "typescript", "ts":
		content, err := mfg.generateTypeScript()
		return content, ".ts", err
	case "descriptor", "desc":
		content, err := mfg.generateDescriptor()
		return content, ".desc", err
	default:
		return nil, "", fmt.Errorf("unsupported format: %s", format)
	}
}

// generateJSONSchema creates a JSON Schema representation
func (mfg *MultiFormatGenerator) generateJSONSchema() ([]byte, error) {
	schema := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-07/schema#",
		"title":       mfg.getPackageName(),
		"description": fmt.Sprintf("JSON Schema for %s protobuf package", mfg.getPackageName()),
		"type":        "object",
		"definitions": make(map[string]interface{}),
		"properties":  make(map[string]interface{}),
	}

	definitions := schema["definitions"].(map[string]interface{})
	properties := schema["properties"].(map[string]interface{})

	// Generate definitions for structs
	for _, structInfo := range mfg.generator.ctx.Structs {
		if mfg.generator.shouldSkipStruct(structInfo) {
			continue
		}

		structSchema := mfg.generateStructJSONSchema(structInfo)
		definitions[structInfo.Name] = structSchema
		properties[strings.ToLower(structInfo.Name)] = map[string]interface{}{
			"$ref": fmt.Sprintf("#/definitions/%s", structInfo.Name),
		}
	}

	// Generate definitions for enums
	for _, enumInfo := range mfg.generator.ctx.Enums {
		if mfg.generator.shouldSkipEnum(enumInfo) {
			continue
		}

		enumSchema := mfg.generateEnumJSONSchema(enumInfo)
		definitions[enumInfo.Name] = enumSchema
	}

	return json.MarshalIndent(schema, "", "  ")
}

// generateStructJSONSchema creates a JSON Schema for a struct
func (mfg *MultiFormatGenerator) generateStructJSONSchema(structInfo *parser.StructInfo) map[string]interface{} {
	// Handle type aliases - generate a reference instead of a full schema
	if structInfo.IsAliasInstantiation {
		aliasTarget := structInfo.AliasTarget
		// For JSON Schema, we return a reference to the base type
		return map[string]interface{}{
			"$ref":        fmt.Sprintf("#/definitions/%s", aliasTarget),
			"description": mfg.generator.getStructDescription(structInfo),
		}
	}

	schema := map[string]interface{}{
		"type":        "object",
		"description": mfg.generator.getStructDescription(structInfo),
		"properties":  make(map[string]interface{}),
	}

	properties := schema["properties"].(map[string]interface{})
	required := make([]string, 0)

	for _, field := range structInfo.Fields {
		if skip, _ := mfg.generator.shouldSkipFieldForMessage(field, structInfo.Name); skip {
			continue
		}

		fieldSchema := mfg.generateFieldJSONSchema(field)
		fieldName := mfg.generator.getFieldJSONName(field)
		properties[fieldName] = fieldSchema

		// Check if field is required (not a pointer and not optional)
		fieldTypeStr := mfg.typeExprToString(field.Type)
		if !strings.HasPrefix(fieldTypeStr, "*") && !mfg.isOptionalField(field) {
			required = append(required, fieldName)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// typeExprToString converts an AST expression to a string representation
func (mfg *MultiFormatGenerator) typeExprToString(t ast.Expr) string {
	switch v := t.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.StarExpr:
		return "*" + mfg.typeExprToString(v.X)
	case *ast.ArrayType:
		return "[]" + mfg.typeExprToString(v.Elt)
	case *ast.SelectorExpr:
		return mfg.typeExprToString(v.X) + "." + v.Sel.Name
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", mfg.typeExprToString(v.Key), mfg.typeExprToString(v.Value))
	case *ast.IndexExpr:
		// Handle generic types like Response[Human]
		baseType := mfg.typeExprToString(v.X)
		indexType := mfg.typeExprToString(v.Index)
		return fmt.Sprintf("%s[%s]", baseType, indexType)
	case *ast.IndexListExpr:
		// Handle generic types with multiple type parameters like Map[K, V]
		baseType := mfg.typeExprToString(v.X)
		var params []string
		for _, idx := range v.Indices {
			params = append(params, mfg.typeExprToString(idx))
		}
		return fmt.Sprintf("%s[%s]", baseType, strings.Join(params, ", "))
	}
	return "unknown"
}

// generateFieldJSONSchema creates a JSON Schema for a field
func (mfg *MultiFormatGenerator) generateFieldJSONSchema(field *parser.FieldInfo) map[string]interface{} {
	schema := map[string]interface{}{
		"description": mfg.generator.getFieldDescription(field),
	}

	fieldTypeStr := mfg.typeExprToString(field.Type)

	// Handle array types
	if strings.HasPrefix(fieldTypeStr, "[]") {
		schema["type"] = "array"
		itemType := strings.TrimPrefix(fieldTypeStr, "[]")
		schema["items"] = mfg.getJSONSchemaType(itemType)
		return schema
	}

	// Handle map types
	if strings.HasPrefix(fieldTypeStr, "map[") {
		schema["type"] = "object"
		schema["additionalProperties"] = mfg.getJSONSchemaType(mfg.extractMapValueType(fieldTypeStr))
		return schema
	}

	// Handle basic and custom types
	jsonType := mfg.getJSONSchemaType(fieldTypeStr)
	for k, v := range jsonType {
		schema[k] = v
	}

	return schema
}

// generateEnumJSONSchema creates a JSON Schema for an enum
func (mfg *MultiFormatGenerator) generateEnumJSONSchema(enumInfo *parser.EnumInfo) map[string]interface{} {
	schema := map[string]interface{}{
		"type":        "string",
		"description": mfg.generator.getEnumDescription(enumInfo),
		"enum":        make([]string, 0),
	}

	enumValues := schema["enum"].([]string)
	for _, value := range enumInfo.Values {
		enumValues = append(enumValues, value.Name)
	}
	schema["enum"] = enumValues

	return schema
}

// getJSONSchemaType converts Go types to JSON Schema types
func (mfg *MultiFormatGenerator) getJSONSchemaType(goType string) map[string]interface{} {
	// Remove pointer prefix
	goType = strings.TrimPrefix(goType, "*")

	switch goType {
	case "string":
		return map[string]interface{}{"type": "string"}
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return map[string]interface{}{"type": "integer"}
	case "float32", "float64":
		return map[string]interface{}{"type": "number"}
	case "bool":
		return map[string]interface{}{"type": "boolean"}
	case "time.Time":
		return map[string]interface{}{
			"type":   "string",
			"format": "date-time",
		}
	default:
		// Custom type - reference to definition
		if mfg.isCustomType(goType) {
			return map[string]interface{}{
				"$ref": fmt.Sprintf("#/definitions/%s", goType),
			}
		}
		return map[string]interface{}{"type": "string"}
	}
}

// Helper methods
func (mfg *MultiFormatGenerator) getPackageName() string {
	if mfg.generator.formatGen.config.Package != "" {
		return mfg.generator.formatGen.config.Package
	}
	return "generated"
}

func (mfg *MultiFormatGenerator) isOptionalField(field *parser.FieldInfo) bool {
	// Check if field has proto optional annotation
	for _, ann := range field.Annotations {
		if ann.Name == "field" || ann.Name == "proto.field" {
			if optional, exists := ann.GetParamBool("optional"); exists && optional {
				return true
			}
		}
	}
	return false
}

func (mfg *MultiFormatGenerator) isCustomType(typeName string) bool {
	// Check if it's one of our defined structs or enums
	for _, s := range mfg.generator.ctx.Structs {
		if s.Name == typeName {
			return true
		}
	}
	for _, e := range mfg.generator.ctx.Enums {
		if e.Name == typeName {
			return true
		}
	}
	return false
}

func (mfg *MultiFormatGenerator) extractMapValueType(mapType string) string {
	// Extract value type from map[key]value
	parts := strings.Split(mapType, "]")
	if len(parts) > 1 {
		return parts[1]
	}
	return "string"
}

// Placeholder methods for other format generators
func (mfg *MultiFormatGenerator) generateMarkdown() ([]byte, error) {
	var out strings.Builder

	out.WriteString(fmt.Sprintf("# %s Protocol Buffer Documentation\n\n", mfg.getPackageName()))
	out.WriteString("Generated from Go structs with protobuf annotations.\n\n")

	// Table of Contents
	out.WriteString("## Table of Contents\n\n")
	out.WriteString("- [Messages](#messages)\n")
	out.WriteString("- [Enums](#enums)\n")
	if mfg.hasServices() {
		out.WriteString("- [Services](#services)\n")
	}
	out.WriteString("\n")

	// Messages
	out.WriteString("## Messages\n\n")
	for _, structInfo := range mfg.generator.ctx.Structs {
		if mfg.generator.shouldSkipStruct(structInfo) {
			continue
		}
		mfg.generateMessageMarkdown(&out, structInfo)
	}

	// Enums
	out.WriteString("## Enums\n\n")
	for _, enumInfo := range mfg.generator.ctx.Enums {
		if mfg.generator.shouldSkipEnum(enumInfo) {
			continue
		}
		mfg.generateEnumMarkdown(&out, enumInfo)
	}

	// Services
	if mfg.hasServices() {
		out.WriteString("## Services\n\n")
		for _, serviceInfo := range mfg.generator.ctx.Interfaces {
			mfg.generateServiceMarkdown(&out, serviceInfo)
		}
	}

	return []byte(out.String()), nil
}

func (mfg *MultiFormatGenerator) generateMessageMarkdown(out *strings.Builder, structInfo *parser.StructInfo) {
	fmt.Fprintf(out, "### %s\n\n", structInfo.Name)

	desc := mfg.generator.getStructDescription(structInfo)
	if desc != "" {
		fmt.Fprintf(out, "%s\n\n", desc)
	}

	// Handle type aliases - generate type alias documentation instead of table
	if structInfo.IsAliasInstantiation {
		aliasType := structInfo.AliasTarget
		if len(structInfo.AliasTypeArgs) > 0 {
			aliasType = fmt.Sprintf("%s[%s]", aliasType, strings.Join(structInfo.AliasTypeArgs, ", "))
		}
		fmt.Fprintf(out, "Type alias for `%s`.\n\n", aliasType)
		return
	}

	out.WriteString("| Field | Type | Description |\n")
	out.WriteString("|-------|------|-------------|\n")

	for _, field := range structInfo.Fields {
		if skip, _ := mfg.generator.shouldSkipFieldForMessage(field, structInfo.Name); skip {
			continue
		}

		fieldName := mfg.generator.getFieldJSONName(field)
		fieldTypeStr := mfg.typeExprToString(field.Type)
		fieldType := mfg.getMarkdownType(fieldTypeStr)
		fieldDesc := mfg.generator.getFieldDescription(field)

		fmt.Fprintf(out, "| %s | %s | %s |\n", fieldName, fieldType, fieldDesc)
	}
	out.WriteString("\n")
}

func (mfg *MultiFormatGenerator) generateEnumMarkdown(out *strings.Builder, enumInfo *parser.EnumInfo) {
	fmt.Fprintf(out, "### %s\n\n", enumInfo.Name)

	desc := mfg.generator.getEnumDescription(enumInfo)
	if desc != "" {
		fmt.Fprintf(out, "%s\n\n", desc)
	}

	out.WriteString("| Name | Value | Description |\n")
	out.WriteString("|------|-------|-------------|\n")

	for _, value := range enumInfo.Values {
		valueDesc := ""
		for _, ann := range value.Annotations {
			if ann.Name == "description" || ann.Name == "doc" {
				if desc, exists := ann.GetParamValue("value"); exists {
					valueDesc = desc
					break
				}
			}
		}
		fmt.Fprintf(out, "| %s | %d | %s |\n", value.Name, value.Value, valueDesc)
	}
	out.WriteString("\n")
}

func (mfg *MultiFormatGenerator) generateServiceMarkdown(out *strings.Builder, serviceInfo *parser.InterfaceInfo) {
	fmt.Fprintf(out, "### %s\n\n", serviceInfo.Name)

	desc := mfg.generator.getInterfaceDescription(serviceInfo)
	if desc != "" {
		fmt.Fprintf(out, "%s\n\n", desc)
	}

	out.WriteString("| Method | Input | Output | Description |\n")
	out.WriteString("|--------|-------|--------|-------------|\n")

	for _, method := range serviceInfo.Methods {
		methodDesc := ""
		// Extract description from annotations if available
		for _, ann := range method.Annotations {
			if ann.Name == "description" || ann.Name == "rpc" {
				if desc, exists := ann.GetParamValue("description"); exists {
					methodDesc = desc
					break
				}
			}
		}

		inputType := mfg.getMethodInputType(method)
		outputType := mfg.getMethodOutputType(method)

		fmt.Fprintf(out, "| %s | %s | %s | %s |\n", method.Name, inputType, outputType, methodDesc)
	}
	out.WriteString("\n")
}

func (mfg *MultiFormatGenerator) generateTypeScript() ([]byte, error) {
	var out strings.Builder

	out.WriteString("// TypeScript definitions generated from protobuf annotations\n")
	out.WriteString("// DO NOT EDIT\n\n")

	// Generate enum types
	for _, enumInfo := range mfg.generator.ctx.Enums {
		if mfg.generator.shouldSkipEnum(enumInfo) {
			continue
		}
		mfg.generateEnumTypeScript(&out, enumInfo)
	}

	// Generate interface types
	for _, structInfo := range mfg.generator.ctx.Structs {
		if mfg.generator.shouldSkipStruct(structInfo) {
			continue
		}
		mfg.generateInterfaceTypeScript(&out, structInfo)
	}

	return []byte(out.String()), nil
}

func (mfg *MultiFormatGenerator) generateEnumTypeScript(out *strings.Builder, enumInfo *parser.EnumInfo) {
	desc := mfg.generator.getEnumDescription(enumInfo)
	if desc != "" {
		fmt.Fprintf(out, "/** %s */\n", desc)
	}

	fmt.Fprintf(out, "export enum %s {\n", enumInfo.Name)
	for _, value := range enumInfo.Values {
		fmt.Fprintf(out, "  %s = %d,\n", value.Name, value.Value)
	}
	out.WriteString("}\n\n")
}

func (mfg *MultiFormatGenerator) generateInterfaceTypeScript(out *strings.Builder, structInfo *parser.StructInfo) {
	desc := mfg.generator.getStructDescription(structInfo)
	if desc != "" {
		fmt.Fprintf(out, "/** %s */\n", desc)
	}

	// Handle type aliases - generate type alias instead of interface
	if structInfo.IsAliasInstantiation {
		aliasType := structInfo.AliasTarget
		if len(structInfo.AliasTypeArgs) > 0 {
			typeArgsTS := make([]string, len(structInfo.AliasTypeArgs))
			for i, arg := range structInfo.AliasTypeArgs {
				typeArgsTS[i] = mfg.getTypeScriptType(arg)
			}
			aliasType = fmt.Sprintf("%s<%s>", aliasType, strings.Join(typeArgsTS, ", "))
		}
		fmt.Fprintf(out, "export type %s = %s;\n\n", structInfo.Name, aliasType)
		return
	}

	// Handle generic types
	interfaceName := structInfo.Name
	genericParams := ""
	if len(structInfo.TypeParams) > 0 {
		genericParams = "<" + strings.Join(structInfo.TypeParams, ", ") + ">"
	}

	fmt.Fprintf(out, "export interface %s%s {\n", interfaceName, genericParams)
	for _, field := range structInfo.Fields {
		if skip, _ := mfg.generator.shouldSkipFieldForMessage(field, structInfo.Name); skip {
			continue
		}

		fieldName := mfg.generator.getFieldJSONName(field)
		fieldTypeStr := mfg.typeExprToString(field.Type)
		fieldType := mfg.getTypeScriptType(fieldTypeStr)
		optional := ""
		if strings.HasPrefix(fieldTypeStr, "*") || mfg.isOptionalField(field) {
			optional = "?"
		}

		fieldDesc := mfg.generator.getFieldDescription(field)
		if fieldDesc != "" {
			fmt.Fprintf(out, "  /** %s */\n", fieldDesc)
		}

		fmt.Fprintf(out, "  %s%s: %s;\n", fieldName, optional, fieldType)
	}
	out.WriteString("}\n\n")
}

func (mfg *MultiFormatGenerator) generateDescriptor() ([]byte, error) {
	// This would generate a binary protobuf descriptor
	// For now, return a placeholder
	return []byte("Binary protobuf descriptor - not implemented yet"), nil
}

// Helper methods for type conversions
func (mfg *MultiFormatGenerator) getMarkdownType(goType string) string {
	goType = strings.TrimPrefix(goType, "*")
	if strings.HasPrefix(goType, "[]") {
		return fmt.Sprintf("[]%s", mfg.getMarkdownType(strings.TrimPrefix(goType, "[]")))
	}
	if strings.HasPrefix(goType, "map[") {
		return fmt.Sprintf("map[%s]", mfg.extractMapValueType(goType))
	}
	return fmt.Sprintf("`%s`", goType)
}

func (mfg *MultiFormatGenerator) getTypeScriptType(goType string) string {
	goType = strings.TrimPrefix(goType, "*")

	if strings.HasPrefix(goType, "[]") {
		itemType := mfg.getTypeScriptType(strings.TrimPrefix(goType, "[]"))
		return fmt.Sprintf("%s[]", itemType)
	}

	if strings.HasPrefix(goType, "map[") {
		valueType := mfg.getTypeScriptType(mfg.extractMapValueType(goType))
		return fmt.Sprintf("Record<string, %s>", valueType)
	}

	// Handle generic types like Response[Human] or Map[string, int]
	if strings.Contains(goType, "[") && strings.Contains(goType, "]") {
		openBracket := strings.Index(goType, "[")
		closeBracket := strings.LastIndex(goType, "]")

		if openBracket > 0 && closeBracket > openBracket {
			baseType := goType[:openBracket]
			typeParams := goType[openBracket+1 : closeBracket]

			// Convert base type
			baseTypeTS := mfg.getTypeScriptType(baseType)

			// Handle type parameters
			var paramsParts []string
			if strings.Contains(typeParams, ",") {
				// Multiple parameters like "K, V"
				for _, param := range strings.Split(typeParams, ",") {
					paramsParts = append(paramsParts, mfg.getTypeScriptType(strings.TrimSpace(param)))
				}
			} else {
				// Single parameter
				paramsParts = append(paramsParts, mfg.getTypeScriptType(strings.TrimSpace(typeParams)))
			}

			return fmt.Sprintf("%s<%s>", baseTypeTS, strings.Join(paramsParts, ", "))
		}
	}

	// Handle generic type parameters (single uppercase letters)
	if len(goType) == 1 && strings.ToUpper(goType) == goType && goType >= "A" && goType <= "Z" {
		return goType // Keep generic type parameter as-is
	}

	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "number"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "time.Time":
		return "string | Date"
	case "interface{}", "any":
		return "any"
	default:
		// Handle custom types and enums
		if strings.Contains(goType, ".") {
			parts := strings.Split(goType, ".")
			return parts[len(parts)-1]
		}
		return goType
	}
}

func (mfg *MultiFormatGenerator) hasServices() bool {
	return len(mfg.generator.ctx.Interfaces) > 0
}

func (mfg *MultiFormatGenerator) getMethodInputType(method *parser.MethodInfo) string {
	if len(method.Params) > 0 {
		return method.Params[0].Name
	}
	return "void"
}

func (mfg *MultiFormatGenerator) getMethodOutputType(method *parser.MethodInfo) string {
	if len(method.Results) > 0 {
		return method.Results[0].Name
	}
	return "void"
}
