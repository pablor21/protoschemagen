package plugin

// WE DO NOT SUPPORT THIS FOR NOW

// import (
// 	"fmt"
// 	"regexp"
// 	"strconv"
// 	"strings"

// 	"github.com/pablor21/gonnotation/annotations"
// 	"github.com/pablor21/gonnotation/parser"
// )

// // ValidationRule represents a validation rule for a field
// type ValidationRule struct {
// 	Type      string      // required, min, max, pattern, etc.
// 	Value     interface{} // the validation value
// 	Message   string      // custom error message
// 	Condition string      // condition when rule applies
// }

// // ValidationContext holds validation state
// type ValidationContext struct {
// 	Rules  map[string][]ValidationRule // field name -> rules
// 	Errors []ValidationError
// 	Logger parser.Logger
// }

// // ValidationError represents a validation error
// type ValidationError struct {
// 	Field   string
// 	Rule    string
// 	Message string
// 	Value   interface{}
// }

// // NewValidationContext creates a new validation context
// func NewValidationContext(logger parser.Logger) *ValidationContext {
// 	return &ValidationContext{
// 		Rules:  make(map[string][]ValidationRule),
// 		Errors: []ValidationError{},
// 		Logger: logger,
// 	}
// }

// // ExtractValidationRules extracts validation rules from annotations
// func (vc *ValidationContext) ExtractValidationRules(structs []*parser.StructInfo) {
// 	for _, structInfo := range structs {
// 		for _, field := range structInfo.Fields {
// 			fieldName := field.Name
// 			rules := vc.extractFieldValidationRules(field)
// 			if len(rules) > 0 {
// 				vc.Rules[fieldName] = rules
// 			}
// 		}
// 	}
// }

// // extractFieldValidationRules extracts validation rules from field annotations
// func (vc *ValidationContext) extractFieldValidationRules(field *parser.FieldInfo) []ValidationRule {
// 	var rules []ValidationRule

// 	for _, ann := range field.Annotations {
// 		if strings.ToLower(ann.Name) == "validate" || strings.HasSuffix(strings.ToLower(ann.Name), ".validate") {
// 			rules = append(rules, vc.parseValidationAnnotation(ann)...)
// 		}
// 	}

// 	return rules
// }

// // parseValidationAnnotation parses a validation annotation into rules
// func (vc *ValidationContext) parseValidationAnnotation(ann annotations.Annotation) []ValidationRule {
// 	var rules []ValidationRule

// 	// Required validation
// 	if required, exists := ann.GetParamBool("required"); exists && required {
// 		rules = append(rules, ValidationRule{
// 			Type:    "required",
// 			Value:   true,
// 			Message: "Field is required",
// 		})
// 	}

// 	// Min value validation
// 	if min, exists := ann.GetParamValue("min"); exists {
// 		if minVal, err := strconv.ParseFloat(min, 64); err == nil {
// 			rules = append(rules, ValidationRule{
// 				Type:    "min",
// 				Value:   minVal,
// 				Message: fmt.Sprintf("Value must be at least %v", minVal),
// 			})
// 		}
// 	}

// 	// Max value validation
// 	if max, exists := ann.GetParamValue("max"); exists {
// 		if maxVal, err := strconv.ParseFloat(max, 64); err == nil {
// 			rules = append(rules, ValidationRule{
// 				Type:    "max",
// 				Value:   maxVal,
// 				Message: fmt.Sprintf("Value must be at most %v", maxVal),
// 			})
// 		}
// 	}

// 	// Min length validation
// 	if minLength, exists := ann.GetParamValue("min_length"); exists {
// 		if minLenVal, err := strconv.Atoi(minLength); err == nil {
// 			rules = append(rules, ValidationRule{
// 				Type:    "min_length",
// 				Value:   minLenVal,
// 				Message: fmt.Sprintf("Length must be at least %d", minLenVal),
// 			})
// 		}
// 	}

// 	// Max length validation
// 	if maxLength, exists := ann.GetParamValue("max_length"); exists {
// 		if maxLenVal, err := strconv.Atoi(maxLength); err == nil {
// 			rules = append(rules, ValidationRule{
// 				Type:    "max_length",
// 				Value:   maxLenVal,
// 				Message: fmt.Sprintf("Length must be at most %d", maxLenVal),
// 			})
// 		}
// 	}

// 	// Pattern validation
// 	if pattern, exists := ann.GetParamValue("pattern"); exists {
// 		if _, err := regexp.Compile(pattern); err == nil {
// 			rules = append(rules, ValidationRule{
// 				Type:    "pattern",
// 				Value:   pattern,
// 				Message: fmt.Sprintf("Value must match pattern: %s", pattern),
// 			})
// 		} else {
// 			vc.Errors = append(vc.Errors, ValidationError{
// 				Rule:    "pattern",
// 				Message: fmt.Sprintf("Invalid regex pattern: %s", pattern),
// 				Value:   pattern,
// 			})
// 		}
// 	}

// 	// Email validation
// 	if email, exists := ann.GetParamBool("email"); exists && email {
// 		rules = append(rules, ValidationRule{
// 			Type:    "email",
// 			Value:   true,
// 			Message: "Value must be a valid email address",
// 		})
// 	}

// 	// URI validation
// 	if uri, exists := ann.GetParamBool("uri"); exists && uri {
// 		rules = append(rules, ValidationRule{
// 			Type:    "uri",
// 			Value:   true,
// 			Message: "Value must be a valid URI",
// 		})
// 	}

// 	// UUID validation
// 	if uuid, exists := ann.GetParamBool("uuid"); exists && uuid {
// 		rules = append(rules, ValidationRule{
// 			Type:    "uuid",
// 			Value:   true,
// 			Message: "Value must be a valid UUID",
// 		})
// 	}

// 	// In validation (allowed values)
// 	if in, exists := ann.GetParamValue("in"); exists {
// 		values := strings.Split(in, ",")
// 		for i, v := range values {
// 			values[i] = strings.TrimSpace(v)
// 		}
// 		rules = append(rules, ValidationRule{
// 			Type:    "in",
// 			Value:   values,
// 			Message: fmt.Sprintf("Value must be one of: %s", strings.Join(values, ", ")),
// 		})
// 	}

// 	// Not in validation (disallowed values)
// 	if notIn, exists := ann.GetParamValue("not_in"); exists {
// 		values := strings.Split(notIn, ",")
// 		for i, v := range values {
// 			values[i] = strings.TrimSpace(v)
// 		}
// 		rules = append(rules, ValidationRule{
// 			Type:    "not_in",
// 			Value:   values,
// 			Message: fmt.Sprintf("Value must not be one of: %s", strings.Join(values, ", ")),
// 		})
// 	}

// 	return rules
// }

// // GenerateValidationCode generates validation code for protobuf messages
// func (vc *ValidationContext) GenerateValidationCode(structInfo *parser.StructInfo) string {
// 	var code strings.Builder

// 	code.WriteString(fmt.Sprintf("// Validate%s validates the %s message\n", structInfo.Name, structInfo.Name))
// 	code.WriteString(fmt.Sprintf("func (m *%s) Validate() error {\n", structInfo.Name))

// 	for _, field := range structInfo.Fields {
// 		if rules, exists := vc.Rules[field.Name]; exists {
// 			for _, rule := range rules {
// 				code.WriteString(vc.generateValidationCodeForRule(field, rule))
// 			}
// 		}
// 	}

// 	code.WriteString("\treturn nil\n")
// 	code.WriteString("}\n\n")

// 	return code.String()
// }

// // generateValidationCodeForRule generates validation code for a specific rule
// func (vc *ValidationContext) generateValidationCodeForRule(field *parser.FieldInfo, rule ValidationRule) string {
// 	fieldName := field.Name
// 	var code strings.Builder

// 	switch rule.Type {
// 	case "required":
// 		if strings.Contains(field.Type, "*") {
// 			code.WriteString(fmt.Sprintf("\tif m.%s == nil {\n", fieldName))
// 			code.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: %s\")\n", fieldName, rule.Message))
// 			code.WriteString("\t}\n")
// 		} else if field.Type == "string" {
// 			code.WriteString(fmt.Sprintf("\tif m.%s == \"\" {\n", fieldName))
// 			code.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: %s\")\n", fieldName, rule.Message))
// 			code.WriteString("\t}\n")
// 		}

// 	case "min":
// 		minVal := rule.Value.(float64)
// 		code.WriteString(fmt.Sprintf("\tif float64(m.%s) < %v {\n", fieldName, minVal))
// 		code.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: %s\")\n", fieldName, rule.Message))
// 		code.WriteString("\t}\n")

// 	case "max":
// 		maxVal := rule.Value.(float64)
// 		code.WriteString(fmt.Sprintf("\tif float64(m.%s) > %v {\n", fieldName, maxVal))
// 		code.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: %s\")\n", fieldName, rule.Message))
// 		code.WriteString("\t}\n")

// 	case "min_length":
// 		minLen := rule.Value.(int)
// 		if field.Type == "string" {
// 			code.WriteString(fmt.Sprintf("\tif len(m.%s) < %d {\n", fieldName, minLen))
// 		} else if strings.HasPrefix(field.Type, "[]") {
// 			code.WriteString(fmt.Sprintf("\tif len(m.%s) < %d {\n", fieldName, minLen))
// 		}
// 		code.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: %s\")\n", fieldName, rule.Message))
// 		code.WriteString("\t}\n")

// 	case "max_length":
// 		maxLen := rule.Value.(int)
// 		if field.Type == "string" {
// 			code.WriteString(fmt.Sprintf("\tif len(m.%s) > %d {\n", fieldName, maxLen))
// 		} else if strings.HasPrefix(field.Type, "[]") {
// 			code.WriteString(fmt.Sprintf("\tif len(m.%s) > %d {\n", fieldName, maxLen))
// 		}
// 		code.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: %s\")\n", fieldName, rule.Message))
// 		code.WriteString("\t}\n")

// 	case "pattern":
// 		pattern := rule.Value.(string)
// 		code.WriteString(fmt.Sprintf("\tif matched, _ := regexp.MatchString(`%s`, m.%s); !matched {\n", pattern, fieldName))
// 		code.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: %s\")\n", fieldName, rule.Message))
// 		code.WriteString("\t}\n")

// 	case "email":
// 		emailPattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
// 		code.WriteString(fmt.Sprintf("\tif matched, _ := regexp.MatchString(`%s`, m.%s); !matched {\n", emailPattern, fieldName))
// 		code.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: %s\")\n", fieldName, rule.Message))
// 		code.WriteString("\t}\n")

// 	case "uri":
// 		code.WriteString(fmt.Sprintf("\tif _, err := url.Parse(m.%s); err != nil {\n", fieldName))
// 		code.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: %s\")\n", fieldName, rule.Message))
// 		code.WriteString("\t}\n")

// 	case "uuid":
// 		uuidPattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
// 		code.WriteString(fmt.Sprintf("\tif matched, _ := regexp.MatchString(`%s`, m.%s); !matched {\n", uuidPattern, fieldName))
// 		code.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: %s\")\n", fieldName, rule.Message))
// 		code.WriteString("\t}\n")

// 	case "in":
// 		values := rule.Value.([]string)
// 		code.WriteString("\tallowed := []string{")
// 		for i, v := range values {
// 			if i > 0 {
// 				code.WriteString(", ")
// 			}
// 			code.WriteString(fmt.Sprintf(`"%s"`, v))
// 		}
// 		code.WriteString("}\n")
// 		code.WriteString(fmt.Sprintf("\tfound := false\n"))
// 		code.WriteString(fmt.Sprintf("\tfor _, v := range allowed {\n"))
// 		code.WriteString(fmt.Sprintf("\t\tif m.%s == v {\n", fieldName))
// 		code.WriteString(fmt.Sprintf("\t\t\tfound = true\n"))
// 		code.WriteString(fmt.Sprintf("\t\t\tbreak\n"))
// 		code.WriteString(fmt.Sprintf("\t\t}\n"))
// 		code.WriteString(fmt.Sprintf("\t}\n"))
// 		code.WriteString(fmt.Sprintf("\tif !found {\n"))
// 		code.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s: %s\")\n", fieldName, rule.Message))
// 		code.WriteString("\t}\n")
// 	}

// 	return code.String()
// }

// // GetValidationImports returns imports needed for validation code
// func (vc *ValidationContext) GetValidationImports() []string {
// 	imports := []string{"fmt"}

// 	// Check if we need regex
// 	needsRegex := false
// 	needsURL := false

// 	for _, rules := range vc.Rules {
// 		for _, rule := range rules {
// 			switch rule.Type {
// 			case "pattern", "email", "uuid":
// 				needsRegex = true
// 			case "uri":
// 				needsURL = true
// 			}
// 		}
// 	}

// 	if needsRegex {
// 		imports = append(imports, "regexp")
// 	}
// 	if needsURL {
// 		imports = append(imports, "net/url")
// 	}

// 	return imports
// }

// // HasValidationRules returns true if any validation rules are defined
// func (vc *ValidationContext) HasValidationRules() bool {
// 	return len(vc.Rules) > 0
// }
