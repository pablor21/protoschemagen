package plugin

import (
	"fmt"
)

// generateServiceAdapter generates the gRPC service adapter that bridges original types
func (g *StubGenerator) generateServiceAdapter() error {
	// Get template configuration
	templateConfig := g.getTemplateConfig()
	templateData := g.prepareTemplateData(templateConfig)

	// Use template-specific imports
	templateData.PackageImports = g.getImportsForTemplate("adapter", templateData.PackageImports)

	// Execute adapter template
	templateNames := templateConfig.GetTemplateNames()
	content, err := g.executeTemplateByName(templateNames["adapter"], templateData)
	if err != nil {
		return fmt.Errorf("failed to generate adapter from template: %w", err)
	}

	return g.writeFile("adapter.go", content)
}

// // generateAdapterImports generates imports for adapter files
// func (g *StubGenerator) generateAdapterImports() string {
// 	var imports strings.Builder
// 	imports.WriteString("import (\n")
// 	imports.WriteString("\t\"context\"\n")
// 	imports.WriteString("\t\"io\"\n\n")

// 	// Add gRPC imports
// 	imports.WriteString("\t\"google.golang.org/grpc\"\n")

// 	// Add protobuf generated package
// 	imports.WriteString("\tpb \"github.com/example/proto/starwars/v1\"\n")
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

// // generateAdapter generates the service adapter implementation
// func (g *StubGenerator) generateAdapter(service *ServiceInfo) string {
// 	var adapter strings.Builder

// 	adapterName := fmt.Sprintf("%sAdapter", service.Name)

// 	// Generate adapter struct
// 	adapter.WriteString(fmt.Sprintf("// %s wraps the original service and provides gRPC compatibility\n", adapterName))
// 	adapter.WriteString(fmt.Sprintf("type %s struct {\n", adapterName))
// 	adapter.WriteString(fmt.Sprintf("\tpb.Unimplemented%sServer\n", service.Name))
// 	adapter.WriteString(fmt.Sprintf("\tservice %s // Original service interface\n", service.Name))
// 	adapter.WriteString("}\n\n")

// 	// Generate constructor
// 	adapter.WriteString(fmt.Sprintf("// New%s creates a new adapter for the service\n", adapterName))
// 	adapter.WriteString(fmt.Sprintf("func New%s(service %s) *%s {\n", adapterName, service.Name, adapterName))
// 	adapter.WriteString(fmt.Sprintf("\treturn &%s{service: service}\n", adapterName))
// 	adapter.WriteString("}\n\n")

// 	// Generate method implementations
// 	for _, method := range service.Methods {
// 		adapter.WriteString(g.generateAdapterMethod(service, method))
// 	}

// 	return adapter.String()
// }

// generateAdapterMethod generates an individual adapter method
// func (g *StubGenerator) generateAdapterMethod(service *ServiceInfo, method *MethodInfo) string {
// 	var methodImpl strings.Builder

// 	adapterName := fmt.Sprintf("%sAdapter", service.Name)

// 	methodImpl.WriteString(fmt.Sprintf("// %s implements the gRPC %s method\n", method.Name, method.Name))

// 	if method.IsStreaming {
// 		methodImpl.WriteString(g.generateStreamingMethod(adapterName, method))
// 	} else {
// 		methodImpl.WriteString(g.generateUnaryMethod(adapterName, method))
// 	}

// 	methodImpl.WriteString("\n")
// 	return methodImpl.String()
// }

// generateUnaryMethod generates a unary method implementation
// func (g *StubGenerator) generateUnaryMethod(adapterName string, method *MethodInfo) string {
// 	var impl strings.Builder

// 	// Method signature
// 	impl.WriteString(fmt.Sprintf("func (a *%s) %s(ctx context.Context, protoReq *pb.%s) (*pb.%s, error) {\n",
// 		adapterName, method.Name, method.InputType, method.OutputType))

// 	// Convert protobuf request to original type
// 	impl.WriteString("\t// Convert protobuf request to original type\n")
// 	impl.WriteString(fmt.Sprintf("\treq := %sFromProto(protoReq)\n", method.InputType))
// 	impl.WriteString("\n")

// 	// Call original service
// 	impl.WriteString("\t// Call original service with original types\n")
// 	impl.WriteString(fmt.Sprintf("\tresult, err := a.service.%s(req)\n", method.Name))
// 	impl.WriteString("\tif err != nil {\n")
// 	impl.WriteString("\t\treturn nil, err\n")
// 	impl.WriteString("\t}\n\n")

// 	// Convert result back to protobuf
// 	impl.WriteString("\t// Convert result back to protobuf\n")
// 	impl.WriteString(fmt.Sprintf("\treturn %sToProto(result), nil\n", method.OutputType))
// 	impl.WriteString("}")

// 	return impl.String()
// }

// generateStreamingMethod generates a streaming method implementation
// func (g *StubGenerator) generateStreamingMethod(adapterName string, method *MethodInfo) string {
// 	var impl strings.Builder

// 	if method.ClientStream && method.ServerStream {
// 		// Bidirectional streaming
// 		impl.WriteString(g.generateBidirectionalMethod(adapterName, method))
// 	} else if method.ClientStream {
// 		// Client streaming
// 		impl.WriteString(g.generateClientStreamingMethod(adapterName, method))
// 	} else {
// 		// Server streaming
// 		impl.WriteString(g.generateServerStreamingMethod(adapterName, method))
// 	}

// 	return impl.String()
// }

// generateServerStreamingMethod generates a server streaming method
// func (g *StubGenerator) generateServerStreamingMethod(adapterName string, method *MethodInfo) string {
// 	var impl strings.Builder

// 	impl.WriteString(fmt.Sprintf("func (a *%s) %s(protoReq *pb.%s, stream grpc.ServerStreamingServer[pb.%s]) error {\n",
// 		adapterName, method.Name, method.InputType, method.OutputType))

// 	// Convert request
// 	impl.WriteString(fmt.Sprintf("\treq := %sFromProto(protoReq)\n", method.InputType))
// 	impl.WriteString("\n")

// 	// Call original service (assuming it returns a channel)
// 	impl.WriteString(fmt.Sprintf("\tresultChan, err := a.service.%s(req)\n", method.Name))
// 	impl.WriteString("\tif err != nil {\n")
// 	impl.WriteString("\t\treturn err\n")
// 	impl.WriteString("\t}\n\n")

// 	// Stream results
// 	impl.WriteString("\tfor result := range resultChan {\n")
// 	impl.WriteString(fmt.Sprintf("\t\tprotoResult := %sToProto(result)\n", method.OutputType))
// 	impl.WriteString("\t\tif err := stream.Send(protoResult); err != nil {\n")
// 	impl.WriteString("\t\t\treturn err\n")
// 	impl.WriteString("\t\t}\n")
// 	impl.WriteString("\t}\n\n")
// 	impl.WriteString("\treturn nil\n")
// 	impl.WriteString("}")

// 	return impl.String()
// }

// // generateClientStreamingMethod generates a client streaming method
// func (g *StubGenerator) generateClientStreamingMethod(adapterName string, method *MethodInfo) string {
// 	var impl strings.Builder

// 	impl.WriteString(fmt.Sprintf("func (a *%s) %s(stream grpc.ClientStreamingServer[pb.%s, pb.%s]) error {\n",
// 		adapterName, method.Name, method.InputType, method.OutputType))

// 	// Collect all input messages
// 	impl.WriteString(fmt.Sprintf("\tinputChan := make(chan models.%s, 10)\n", method.InputType))
// 	impl.WriteString("\tgo func() {\n")
// 	impl.WriteString("\t\tdefer close(inputChan)\n")
// 	impl.WriteString("\t\tfor {\n")
// 	impl.WriteString("\t\t\tprotoMsg, err := stream.Recv()\n")
// 	impl.WriteString("\t\t\tif err == io.EOF {\n")
// 	impl.WriteString("\t\t\t\treturn\n")
// 	impl.WriteString("\t\t\t}\n")
// 	impl.WriteString("\t\t\tif err != nil {\n")
// 	impl.WriteString("\t\t\t\treturn\n")
// 	impl.WriteString("\t\t\t}\n")
// 	impl.WriteString(fmt.Sprintf("\t\t\tmsg := %sFromProto(protoMsg)\n", method.InputType))
// 	impl.WriteString("\t\t\tinputChan <- msg\n")
// 	impl.WriteString("\t\t}\n")
// 	impl.WriteString("\t}()\n\n")

// 	// Call original service
// 	impl.WriteString(fmt.Sprintf("\tresult, err := a.service.%s(inputChan)\n", method.Name))
// 	impl.WriteString("\tif err != nil {\n")
// 	impl.WriteString("\t\treturn err\n")
// 	impl.WriteString("\t}\n\n")

// 	// Send response
// 	impl.WriteString(fmt.Sprintf("\tprotoResult := %sToProto(result)\n", method.OutputType))
// 	impl.WriteString("\treturn stream.SendAndClose(protoResult)\n")
// 	impl.WriteString("}")

// 	return impl.String()
// }

// // generateBidirectionalMethod generates a bidirectional streaming method
// func (g *StubGenerator) generateBidirectionalMethod(adapterName string, method *MethodInfo) string {
// 	var impl strings.Builder

// 	impl.WriteString(fmt.Sprintf("func (a *%s) %s(stream grpc.BidiStreamingServer[pb.%s, pb.%s]) error {\n",
// 		adapterName, method.Name, method.InputType, method.OutputType))

// 	// Create channels for communication
// 	impl.WriteString(fmt.Sprintf("\tinputChan := make(chan models.%s, 10)\n", method.InputType))
// 	impl.WriteString(fmt.Sprintf("\toutputChan := make(chan models.%s, 10)\n", method.OutputType))
// 	impl.WriteString("\n")

// 	// Start goroutine for receiving messages
// 	impl.WriteString("\tgo func() {\n")
// 	impl.WriteString("\t\tdefer close(inputChan)\n")
// 	impl.WriteString("\t\tfor {\n")
// 	impl.WriteString("\t\t\tprotoMsg, err := stream.Recv()\n")
// 	impl.WriteString("\t\t\tif err == io.EOF {\n")
// 	impl.WriteString("\t\t\t\treturn\n")
// 	impl.WriteString("\t\t\t}\n")
// 	impl.WriteString("\t\t\tif err != nil {\n")
// 	impl.WriteString("\t\t\t\treturn\n")
// 	impl.WriteString("\t\t\t}\n")
// 	impl.WriteString(fmt.Sprintf("\t\t\tmsg := %sFromProto(protoMsg)\n", method.InputType))
// 	impl.WriteString("\t\t\tinputChan <- msg\n")
// 	impl.WriteString("\t\t}\n")
// 	impl.WriteString("\t}()\n\n")

// 	// Start goroutine for sending responses
// 	impl.WriteString("\tgo func() {\n")
// 	impl.WriteString("\t\tfor response := range outputChan {\n")
// 	impl.WriteString(fmt.Sprintf("\t\t\tprotoResponse := %sToProto(response)\n", method.OutputType))
// 	impl.WriteString("\t\t\tif err := stream.Send(protoResponse); err != nil {\n")
// 	impl.WriteString("\t\t\t\treturn\n")
// 	impl.WriteString("\t\t\t}\n")
// 	impl.WriteString("\t\t}\n")
// 	impl.WriteString("\t}()\n\n")

// 	// Call original service
// 	impl.WriteString(fmt.Sprintf("\treturn a.service.%s(inputChan, outputChan)\n", method.Name))
// 	impl.WriteString("}")

// 	return impl.String()
// }

// generateRegistrationHelpers generates helper functions for service registration
func (g *StubGenerator) generateRegistrationHelpers() error {
	// Get template configuration
	templateConfig := g.getTemplateConfig()
	templateData := g.prepareTemplateData(templateConfig)

	// Use template-specific imports
	templateData.PackageImports = g.getImportsForTemplate("registration", templateData.PackageImports)

	// Execute registration template
	templateNames := templateConfig.GetTemplateNames()
	content, err := g.executeTemplateByName(templateNames["registration"], templateData)
	if err != nil {
		return fmt.Errorf("failed to generate registration from template: %w", err)
	}

	return g.writeFile("registration.go", content)
}

// generateRegistrationHelper generates a registration helper for a service
// func (g *StubGenerator) generateRegistrationHelper(service *ServiceInfo) string {
// 	var helper strings.Builder

// 	functionName := fmt.Sprintf("Register%s", service.Name)
// 	adapterName := fmt.Sprintf("%sAdapter", service.Name)

// 	helper.WriteString(fmt.Sprintf("// %s registers the service with a gRPC server using original types\n", functionName))
// 	helper.WriteString(fmt.Sprintf("func %s(server *grpc.Server, service %s) {\n", functionName, service.Name))
// 	helper.WriteString(fmt.Sprintf("\tadapter := New%s(service)\n", adapterName))
// 	helper.WriteString(fmt.Sprintf("\tpb.Register%sServer(server, adapter)\n", service.Name))
// 	helper.WriteString("}\n\n")

// 	return helper.String()
// }

// generateServiceBridge generates the service bridge that converts slice-based to channel-based interface
func (g *StubGenerator) generateServiceBridge() error {
	// Get template configuration
	templateConfig := g.getTemplateConfig()
	templateData := g.prepareTemplateData(templateConfig)

	// Use template-specific imports
	templateData.PackageImports = g.getImportsForTemplate("bridge", templateData.PackageImports)

	// Execute bridge template
	templateNames := templateConfig.GetTemplateNames()
	content, err := g.executeTemplateByName(templateNames["bridge"], templateData)
	if err != nil {
		return fmt.Errorf("failed to generate bridge from template: %w", err)
	}

	return g.writeFile("bridge.go", content)
}
