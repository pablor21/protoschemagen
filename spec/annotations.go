package spec

import "github.com/pablor21/gonnotation/annotations"

// annotationSpecs defines Protobuf annotation specifications
var annotationSpecs = []AnnotationSpec{
	{
		Name:        "package",
		Description: "Package-level protobuf metadata",
		Params: []Param{
			{Name: "name", Types: []string{"string"}, Description: "Protobuf package name", IsRequired: true, IsDefault: true},
			{Name: "go_package", Types: []string{"string"}, Description: "Go package path"},
			{Name: "java_package", Types: []string{"string"}},
			{Name: "java_outer_classname", Types: []string{"string"}},
			{Name: "optimize_for", Types: []string{"string"}, EnumValues: []string{"SPEED", "CODE_SIZE", "LITE_RUNTIME"}},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnPackage, annotations.AnnotationValidOnFile},
	},
	{
		Name:        "message",
		Description: "Defines a protobuf message",
		Params: []Param{
			{Name: "name", Types: []string{"string"}, Description: "Custom message name"},
			{Name: "description", Types: []string{"string"}},
			{Name: "reserved", Types: []string{"bool", "string", "[]string"}, Description: "Mark all fields as reserved: empty or true (all), message name, or array of message names"},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnStruct},
	},
	{
		Name:        "enum",
		Description: "Defines a protobuf enum",
		Params: []Param{
			{Name: "name", Types: []string{"string"}},
			{Name: "description", Types: []string{"string"}},
			{Name: "allow_alias", Types: []string{"bool"}},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnEnum},
	},
	{
		Name:        "enumvalue",
		Description: "Defines metadata for a single enum value",
		Params: []Param{
			{Name: "name", Types: []string{"string"}},
			{Name: "number", Types: []string{"int"}, Description: "Enum value number"},
			{Name: "description", Types: []string{"string"}},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnEnumValue},
	},
	getAnnotationSpecsFromTags("field", "Defines a protobuf field"),
	{
		Name:        "oneof",
		Description: "Groups fields into a oneof",
		Params: []Param{
			{Name: "name", Types: []string{"string"}, IsRequired: true},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnField},
	},
	{
		Name:        "service",
		Description: "Defines a gRPC service",
		Params: []Param{
			{Name: "name", Types: []string{"string"}},
			{Name: "description", Types: []string{"string"}},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnInterface, annotations.AnnotationValidOnStruct},
	},
	{
		Name:        "rpc",
		Description: "Defines an RPC method",
		Params: []Param{
			{Name: "name", Types: []string{"string"}},
			{Name: "description", Types: []string{"string"}},
			{Name: "input", Types: []string{"string"}, Description: "Request message type"},
			{Name: "output", Types: []string{"string"}, Description: "Response message type"},
			{Name: "client_streaming", Types: []string{"bool"}},
			{Name: "server_streaming", Types: []string{"bool"}},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnFunction},
	},
	{
		Name:        "map",
		Description: "Defines a map field",
		Params: []Param{
			{Name: "key", Types: []string{"string"}, IsRequired: true, Description: "Map key type"},
			{Name: "value", Types: []string{"string"}, IsRequired: true, Description: "Map value type"},
			{Name: "number", Types: []string{"int"}, Description: "Field number"},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnField},
	},
	{
		Name:        "import",
		Description: "Imports another proto file",
		Multiple:    true,
		Params: []Param{
			{Name: "path", Types: []string{"string"}, IsRequired: true},
			{Name: "public", Types: []string{"bool"}},
			{Name: "weak", Types: []string{"bool"}},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnPackage, annotations.AnnotationValidOnFile},
	},
	{
		Name:        "reserved",
		Description: "Reserves field numbers or names",
		Multiple:    true,
		Params: []Param{
			{Name: "for", Types: []string{"string", "[]string"}, Description: "Apply reserved only to specific messages: message name or array of names"},
			{Name: "numbers", Types: []string{"string", "[]int"}, Description: "Reserved field numbers or ranges"},
			{Name: "names", Types: []string{"string", "[]string"}, Description: "Reserved field names"},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnStruct},
	},
	{
		Name:        "option",
		Description: "Sets a protobuf option",
		Multiple:    true,
		Params: []Param{
			{Name: "name", Types: []string{"string"}, IsRequired: true},
			{Name: "value", Types: []string{"string"}, IsRequired: true},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnStruct, annotations.AnnotationValidOnField, annotations.AnnotationValidOnEnum, annotations.AnnotationValidOnEnumValue},
	},
	{
		Name:        "extend",
		Description: "Extends an existing message",
		Params: []Param{
			{Name: "message", Types: []string{"string"}, IsRequired: true},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnStruct},
	},
	{
		Name:        "ignore",
		Description: "Skip this type from proto generation",
		Aliases:     []string{"skip", "omit"},
		Params: []Param{
			{Name: "for", Types: []string{"bool", "string", "[]string"}, Description: "Message names to ignore for: true (all), message name, or array of names"},
			{Name: "reserved", Types: []string{"bool", "string", "[]string"}, Description: "Reserve field numbers when ignored: empty or true (all), message name, or array of names"},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnStruct, annotations.AnnotationValidOnField, annotations.AnnotationValidOnEnum},
	},
	{
		Name:        "include",
		Description: "Include this field/type only in specific messages",
		Params: []Param{
			{Name: "for", Types: []string{"bool", "string", "[]string"}, Description: "Message names to include for: true (all), message name, or array of names"},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnField},
	},
}

// getAnnotationSpecsFromTags generates an annotation spec from the tag parameters
func getAnnotationSpecsFromTags(name string, description string) AnnotationSpec {
	spec := AnnotationSpec{
		Name:        name,
		Description: description,
		Multiple:    true, // Allow multiple annotations with different 'for' parameters
		Params: []Param{
			{Name: "for", Types: []string{"string", "[]string"}, Description: "Apply field config only to specific messages: message name or array of names"},
		},
		ValidOn: []ValidOn{annotations.AnnotationValidOnField},
	}

	for _, tagParam := range tagParams {
		param := Param{
			Name:         tagParam.Name,
			Types:        tagParam.Types,
			EnumValues:   tagParam.EnumValues,
			Description:  tagParam.Description,
			Aliases:      tagParam.Aliases,
			DefaultValue: tagParam.DefaultValue,
		}
		spec.Params = append(spec.Params, param)
	}

	return spec
}
