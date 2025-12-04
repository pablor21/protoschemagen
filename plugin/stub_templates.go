package plugin

import (
	"fmt"
)

// generateTypeAdapters generates conversion functions between original and protobuf types
func (g *StubGenerator) generateTypeAdapters() error {
	// Get template configuration
	templateConfig := g.getTemplateConfig()
	templateData := g.prepareTemplateData(templateConfig)

	// Use template-specific imports
	templateData.PackageImports = g.getImportsForTemplate("types", templateData.PackageImports)

	// Execute types template
	templateNames := templateConfig.GetTemplateNames()
	content, err := g.executeTemplateByName(templateNames["types"], templateData)
	if err != nil {
		return fmt.Errorf("failed to generate types from template: %w", err)
	}

	return g.writeFile("types.go", content)
}

// // generateAdapterHeader generates the package header for adapter files
// func (g *StubGenerator) generateAdapterHeader() string {
// 	return `// Package adapter contains auto-generated type adapters
// // Generated from protobuf annotations - DO NOT EDIT
// package adapter

// `
// }

// // generateImports generates the import section for adapter files
// func (g *StubGenerator) generateImports() string {
// 	// Get unique packages from analyzed types
// 	packages := make(map[string]bool)

// 	// Add standard packages
// 	packages["context"] = true
// 	packages["fmt"] = true
// 	packages["./models"] = true

// 	// Add original type packages
// 	for _, typeInfo := range g.originalTypes {
// 		if typeInfo.Package != "" && typeInfo.Package != "./models" {
// 			packages[typeInfo.Package] = true
// 		}
// 	}

// 	// Add protobuf packages - path based on go_package option in protobuf
// 	packages["github.com/example/proto/starwars/v1"] = true
// 	packages["google.golang.org/protobuf/types/known/timestamppb"] = true
// 	packages["google.golang.org/grpc"] = true

// 	var imports strings.Builder
// 	imports.WriteString("import (\n")

// 	for pkg := range packages {
// 		if pkg == "context" || pkg == "fmt" {
// 			imports.WriteString(fmt.Sprintf("\t\"%s\"\n", pkg))
// 		} else if pkg == "github.com/example/proto/starwars/v1" {
// 			imports.WriteString(fmt.Sprintf("\tpb \"%s\"\n", pkg))
// 		} else {
// 			// For other packages, use simple import
// 			imports.WriteString(fmt.Sprintf("\t\"%s\"\n", pkg))
// 		}
// 	}

// 	imports.WriteString(")\n\n")
// 	return imports.String()
// }

// // generateTypeConverter generates conversion functions for a specific type
// func (g *StubGenerator) generateTypeConverter(typeInfo *TypeInfo) string {
// 	var converter strings.Builder

// 	// Generate ToProto function
// 	converter.WriteString(g.generateToProtoFunction(typeInfo))
// 	converter.WriteString("\n")

// 	// Generate FromProto function
// 	converter.WriteString(g.generateFromProtoFunction(typeInfo))
// 	converter.WriteString("\n")

// 	return converter.String()
// }

// // generateToProtoFunction generates the Original -> Protobuf conversion function
// func (g *StubGenerator) generateToProtoFunction(typeInfo *TypeInfo) string {
// 	var fn strings.Builder

// 	originalType := fmt.Sprintf("%s.%s", g.getOriginalPackageAlias(typeInfo.Package), typeInfo.Name)
// 	protoType := fmt.Sprintf("*pb.%s", typeInfo.Name)

// 	fn.WriteString(fmt.Sprintf("// %sToProto converts %s to protobuf %s\n", typeInfo.Name, originalType, protoType))
// 	fn.WriteString(fmt.Sprintf("func %sToProto(orig %s) %s {\n", typeInfo.Name, originalType, protoType))
// 	fn.WriteString(fmt.Sprintf("\tif orig == (%s{}) {\n", originalType))
// 	fn.WriteString("\t\treturn nil\n")
// 	fn.WriteString("\t}\n\n")

// 	fn.WriteString(fmt.Sprintf("\tproto := &pb.%s{\n", typeInfo.Name))

// 	// Generate field assignments
// 	for _, field := range typeInfo.Fields {
// 		fn.WriteString(g.generateFieldToProto(field))
// 	}

// 	fn.WriteString("\t}\n\n")
// 	fn.WriteString("\treturn proto\n")
// 	fn.WriteString("}\n")

// 	return fn.String()
// }

// // generateFromProtoFunction generates the Protobuf -> Original conversion function
// func (g *StubGenerator) generateFromProtoFunction(typeInfo *TypeInfo) string {
// 	var fn strings.Builder

// 	originalType := fmt.Sprintf("%s.%s", g.getOriginalPackageAlias(typeInfo.Package), typeInfo.Name)
// 	protoType := fmt.Sprintf("*pb.%s", typeInfo.Name)

// 	fn.WriteString(fmt.Sprintf("// %sFromProto converts protobuf %s to %s\n", typeInfo.Name, protoType, originalType))
// 	fn.WriteString(fmt.Sprintf("func %sFromProto(proto %s) %s {\n", typeInfo.Name, protoType, originalType))
// 	fn.WriteString("\tif proto == nil {\n")
// 	fn.WriteString(fmt.Sprintf("\t\treturn %s{}\n", originalType))
// 	fn.WriteString("\t}\n\n")

// 	fn.WriteString(fmt.Sprintf("\torig := %s{\n", originalType))

// 	// Generate field assignments
// 	for _, field := range typeInfo.Fields {
// 		fn.WriteString(g.generateFieldFromProto(field))
// 	}

// 	fn.WriteString("\t}\n\n")
// 	fn.WriteString("\treturn orig\n")
// 	fn.WriteString("}\n")

// 	return fn.String()
// }

// // generateFieldToProto generates field assignment for Original -> Protobuf
// func (g *StubGenerator) generateFieldToProto(field *FieldInfo) string {
// 	protoFieldName := field.ProtoName
// 	if protoFieldName == "" {
// 		protoFieldName = g.toSnakeCase(field.Name)
// 	}
// 	protoFieldName = g.capitalizeFirst(protoFieldName)

// 	// Skip fields without proper protobuf mapping
// 	if protoFieldName == "" {
// 		return ""
// 	}

// 	if field.IsRepeated {
// 		return fmt.Sprintf("\t\t%s: orig.%s,\n", protoFieldName, field.Name)
// 	}

// 	// Handle optional fields (pointers)
// 	if strings.HasPrefix(field.Type, "*") {
// 		return fmt.Sprintf("\t\t%s: orig.%s,\n", protoFieldName, field.Name)
// 	}

// 	return fmt.Sprintf("\t\t%s: orig.%s,\n", protoFieldName, field.Name)
// }

// // generateFieldFromProto generates field assignment for Protobuf -> Original
// func (g *StubGenerator) generateFieldFromProto(field *FieldInfo) string {
// 	protoFieldName := field.ProtoName
// 	if protoFieldName == "" {
// 		protoFieldName = g.toSnakeCase(field.Name)
// 	}
// 	protoFieldName = g.capitalizeFirst(protoFieldName)

// 	// Skip fields without proper protobuf mapping
// 	if protoFieldName == "" {
// 		return ""
// 	}

// 	if field.IsRepeated {
// 		return fmt.Sprintf("\t\t%s: proto.%s,\n", field.Name, protoFieldName)
// 	}

// 	// Handle optional fields (pointers)
// 	if strings.HasPrefix(field.Type, "*") {
// 		return fmt.Sprintf("\t\t%s: proto.%s,\n", field.Name, protoFieldName)
// 	}

// 	return fmt.Sprintf("\t\t%s: proto.%s,\n", field.Name, protoFieldName)
// }

// generateOriginalServiceInterface generates service interface using original types
func (g *StubGenerator) generateOriginalServiceInterface() error {
	// Get template configuration
	templateConfig := g.getTemplateConfig()
	templateData := g.prepareTemplateData(templateConfig)

	// Use template-specific imports
	templateData.PackageImports = g.getImportsForTemplate("service", templateData.PackageImports)

	// Execute service template
	templateNames := templateConfig.GetTemplateNames()
	content, err := g.executeTemplateByName(templateNames["service"], templateData)
	if err != nil {
		return fmt.Errorf("failed to generate service from template: %w", err)
	}

	return g.writeFile("service.go", content)
}

// // generateServiceImports generates imports for service files
// func (g *StubGenerator) generateServiceImports() string {
// 	var imports strings.Builder
// 	imports.WriteString("import (\n")
// 	imports.WriteString("\t\"context\"\n")
// 	imports.WriteString("\t\"./models\"\n")

// 	// Add original type packages
// 	packages := make(map[string]bool)
// 	for _, service := range g.services {
// 		if service.Package != "" {
// 			packages[service.Package] = true
// 		}
// 	}

// 	for pkg := range packages {
// 		if pkg != "./models" { // Avoid duplicate
// 			imports.WriteString(fmt.Sprintf("\t\"%s\"\n", pkg))
// 		}
// 	}

// 	imports.WriteString(")\n\n")
// 	return imports.String()
// }

// generateServiceInterface generates the service interface using original types
// func (g *StubGenerator) generateServiceInterface(service *ServiceInfo) string {
// 	var intf strings.Builder

// 	intf.WriteString(fmt.Sprintf("// %s defines the service interface using original Go types\n", service.Name))
// 	intf.WriteString(fmt.Sprintf("type %s interface {\n", service.Name))

// 	for _, method := range service.Methods {
// 		intf.WriteString(g.generateMethodSignature(method))
// 	}

// 	intf.WriteString("}\n\n")
// 	return intf.String()
// }

// generateMethodSignature generates method signature using original types
// func (g *StubGenerator) generateMethodSignature(method *MethodInfo) string {
// 	var signature strings.Builder

// 	signature.WriteString(fmt.Sprintf("\t// %s implements the %s RPC method\n", method.Name, method.Name))

// 	// Qualify type names with models package
// 	inputType := fmt.Sprintf("models.%s", method.InputType)
// 	outputType := fmt.Sprintf("models.%s", method.OutputType)

// 	if method.IsStreaming {
// 		// For streaming methods, we'll use channels or different signatures
// 		if method.ClientStream && method.ServerStream {
// 			// Bidirectional streaming
// 			signature.WriteString(fmt.Sprintf("\t%s(stream chan %s) (chan %s, error)\n",
// 				method.Name, inputType, outputType))
// 		} else if method.ClientStream {
// 			// Client streaming
// 			signature.WriteString(fmt.Sprintf("\t%s(input chan %s) (%s, error)\n",
// 				method.Name, inputType, outputType))
// 		} else {
// 			// Server streaming
// 			signature.WriteString(fmt.Sprintf("\t%s(req %s) (chan %s, error)\n",
// 				method.Name, inputType, outputType))
// 		}
// 	} else {
// 		// Unary method
// 		signature.WriteString(fmt.Sprintf("\t%s(req %s) (%s, error)\n",
// 			method.Name, inputType, outputType))
// 	}

// 	return signature.String()
// }

// // Helper functions
// func (g *StubGenerator) getOriginalPackageAlias(pkg string) string {
// 	// Simple package alias generation
// 	parts := strings.Split(pkg, "/")
// 	if len(parts) > 0 {
// 		return parts[len(parts)-1]
// 	}
// 	return "models"
// }

// func (g *StubGenerator) capitalizeFirst(s string) string {
// 	if len(s) == 0 {
// 		return s
// 	}
// 	return strings.ToUpper(s[:1]) + s[1:]
// }

// generateClient generates the gRPC client wrapper using Go types
func (g *StubGenerator) generateClient() error {
	// Get template configuration
	templateConfig := g.getTemplateConfig()
	templateData := g.prepareTemplateData(templateConfig)

	// Use template-specific imports
	templateData.PackageImports = g.getImportsForTemplate("client", templateData.PackageImports)

	// Execute client template
	templateNames := templateConfig.GetTemplateNames()
	content, err := g.executeTemplateByName(templateNames["client"], templateData)
	if err != nil {
		return fmt.Errorf("failed to generate client from template: %w", err)
	}

	return g.writeFile("client.go", content)
}
