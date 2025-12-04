package plugin

import (
	"fmt"
	"strings"

	"github.com/pablor21/gonnotation/parser"
	"github.com/pablor21/goschemagen"
)

// ProtoValidator validates protobuf-specific schema rules
type ProtoValidator struct {
	logger parser.Logger
}

// NewProtoValidator creates a new protobuf validator
func NewProtoValidator(logger parser.Logger) *ProtoValidator {
	return &ProtoValidator{
		logger: logger,
	}
}

// ValidateContext validates the generation context for protobuf-specific issues
func (pv *ProtoValidator) ValidateContext(ctx *goschemagen.GenerationContext, gen *Generator) []parser.ValidationError {
	var errors []parser.ValidationError

	messageNames := make(map[string]bool)
	fieldNumbersByMessage := make(map[string]map[int]string) // message -> field_number -> field_name

	// Validate messages
	for _, s := range ctx.Structs {
		// Skip if has other type annotation
		if gen.hasOtherTypeAnnotation(s) {
			continue
		}

		messageNamesForStruct := gen.getMessageNames(s)
		if len(messageNamesForStruct) == 0 {
			messageNamesForStruct = []string{gen.getMessageName(s, "")}
		}

		for _, msgName := range messageNamesForStruct {
			// Check for duplicate message names
			if messageNames[msgName] {
				errors = append(errors, parser.ValidationError{
					Location: fmt.Sprintf("struct %s", s.Name),
					Message:  fmt.Sprintf("duplicate message name: %s", msgName),
					Severity: "error",
				})
			}
			messageNames[msgName] = true

			// Validate field numbers for this message
			fieldNumbersByMessage[msgName] = make(map[int]string)
			reservedNumbers := make(map[int]bool)

			// Collect reserved numbers from @proto.reserved annotation
			for _, ann := range s.Annotations {
				name := strings.ToLower(ann.Name)
				if name == "reserved" || strings.HasSuffix(name, ".reserved") {
					// Check if applies to this message
					if forValue, hasFor := ann.GetParamValue("for"); hasFor {
						if forValue != "" && forValue != "*" && forValue != msgName {
							continue
						}
					}

					// Get reserved numbers
					if numbersStr, ok := ann.GetParamValue("numbers"); ok && numbersStr != "" {
						// Parse numbers/ranges
						parts := strings.Split(numbersStr, ",")
						for _, part := range parts {
							part = strings.TrimSpace(part)
							if num, err := fmt.Sscanf(part, "%d", new(int)); err == nil && num == 1 {
								var n int
								if _, err := fmt.Sscanf(part, "%d", &n); err == nil {
									reservedNumbers[n] = true
								}
							}
						}
					}
				}
			}

			for _, f := range s.Fields {
				if f.GoName == "" || len(f.GoName) == 0 || f.GoName[0] < 'A' || f.GoName[0] > 'Z' {
					continue // skip unexported fields
				}

				// Check if field applies to this message
				if !gen.fieldAnnotationAppliesToMessage(f, msgName) {
					continue
				}

				// Check if field is skipped for this message
				isSkipped, _ := gen.shouldSkipFieldForMessage(f, msgName)
				if isSkipped {
					continue
				}

				// Get field number
				num := gen.getFieldNumberFromAnnotation(f)
				if num > 0 {
					// Check for field number conflicts
					if existingField, exists := fieldNumbersByMessage[msgName][num]; exists {
						errors = append(errors, parser.ValidationError{
							Location: fmt.Sprintf("message %s, field %s", msgName, f.Name),
							Message:  fmt.Sprintf("duplicate field number %d (also used by field '%s')", num, existingField),
							Severity: "error",
						})
					}

					// Check if using reserved number
					if reservedNumbers[num] {
						errors = append(errors, parser.ValidationError{
							Location: fmt.Sprintf("message %s, field %s", msgName, f.Name),
							Message:  fmt.Sprintf("field number %d is reserved", num),
							Severity: "error",
						})
					}

					// Validate field number range (protobuf limits)
					if num < 1 || num > 536870911 {
						errors = append(errors, parser.ValidationError{
							Location: fmt.Sprintf("message %s, field %s", msgName, f.Name),
							Message:  fmt.Sprintf("field number %d is out of valid range (1-536870911)", num),
							Severity: "error",
						})
					}

					// Check reserved ranges (19000-19999 reserved by protobuf)
					if num >= 19000 && num <= 19999 {
						errors = append(errors, parser.ValidationError{
							Location: fmt.Sprintf("message %s, field %s", msgName, f.Name),
							Message:  fmt.Sprintf("field number %d is in reserved range 19000-19999", num),
							Severity: "error",
						})
					}

					fieldNumbersByMessage[msgName][num] = f.Name
				}
			}
		}
	}

	// Validate enums
	enumNames := make(map[string]bool)
	enumValuesByEnum := make(map[string]map[int]string) // enum -> value_number -> value_name

	for _, e := range ctx.Enums {
		enumName := gen.getEnumName(e)

		// Check for duplicate enum names
		if enumNames[enumName] {
			errors = append(errors, parser.ValidationError{
				Location: fmt.Sprintf("enum %s", e.Name),
				Message:  fmt.Sprintf("duplicate enum name: %s", enumName),
				Severity: "error",
			})
		}
		enumNames[enumName] = true

		// Validate enum values
		enumValuesByEnum[enumName] = make(map[int]string)

		for i, val := range e.Values {
			valueNum := i // Default to index

			// Check for custom number in annotation
			for _, ann := range val.Annotations {
				name := strings.ToLower(ann.Name)
				if name == "enumvalue" || strings.HasSuffix(name, ".enumvalue") {
					if numStr := ann.Params["number"]; numStr != "" {
						_, _ = fmt.Sscanf(numStr, "%d", &valueNum) // Ignore error, keep default if parse fails
					}
				}
			}

			// Check for duplicate enum value numbers
			if existingValue, exists := enumValuesByEnum[enumName][valueNum]; exists {
				errors = append(errors, parser.ValidationError{
					Location: fmt.Sprintf("enum %s, value %s", enumName, val.Name),
					Message:  fmt.Sprintf("duplicate enum value number %d (also used by value '%s')", valueNum, existingValue),
					Severity: "error",
				})
			}

			enumValuesByEnum[enumName][valueNum] = val.Name
		}
	}

	// Validate services
	serviceNames := make(map[string]bool)
	for _, iface := range ctx.Interfaces {
		if gen.shouldGenerateService(iface) {
			serviceName := gen.getServiceName(iface)

			// Check for duplicate service names
			if serviceNames[serviceName] {
				errors = append(errors, parser.ValidationError{
					Location: fmt.Sprintf("interface %s", iface.Name),
					Message:  fmt.Sprintf("duplicate service name: %s", serviceName),
					Severity: "error",
				})
			}
			serviceNames[serviceName] = true
		}
	}

	return errors
}
