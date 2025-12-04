package spec

// tagParams defines Protobuf struct tag parameters
var tagParams = []TagParam{
	{
		Name:        "name",
		Types:       []string{"string"},
		Description: "Override protobuf field name",
		IsDefault:   true,
	},
	{
		Name:        "number",
		Types:       []string{"int"},
		Description: "Field number in protobuf message",
	},
	{
		Name:        "type",
		Types:       []string{"string"},
		Description: "Protobuf type override (int32, int64, string, etc.)",
	},
	{
		Name:        "repeated",
		Types:       []string{"bool"},
		Description: "Mark field as repeated (array)",
	},
	{
		Name:        "optional",
		Types:       []string{"bool"},
		Description: "Mark field as optional (proto3)",
	},
	{
		Name:        "packed",
		Types:       []string{"bool"},
		Description: "Use packed encoding for repeated numeric fields",
	},
	{
		Name:        "deprecated",
		Types:       []string{"bool"},
		Description: "Mark field as deprecated",
	},
	{
		Name:        "json_name",
		Types:       []string{"string"},
		Description: "JSON field name override",
	},
	{
		Name:        "oneof",
		Types:       []string{"string"},
		Description: "Oneof group name",
	},
	{
		Name:        "default",
		Types:       []string{"string"},
		Description: "Default value (proto2)",
	},
	{
		Name:        "for",
		Types:       []string{"string", "[]string"},
		Description: "Apply field config only to specific messages: message name or array of names",
	},
	{
		Name:         "ignore",
		Types:        []string{"null", "bool", "string", "[]string"},
		Description:  "Skip this field: empty or true (all messages), message name, or array of message names",
		Aliases:      []string{"omit", "skip"},
		DefaultValue: "*",
	},
	{
		Name:         "include",
		Types:        []string{"null", "bool", "string", "[]string"},
		Description:  "Include only in messages: empty or true (all), message name, or array of names",
		DefaultValue: "*",
	},
	{
		Name:         "reserved",
		Types:        []string{"null", "bool", "string", "[]string"},
		Description:  "Reserve field number when omitted: empty or true (all), message name, or array of names",
		DefaultValue: "*",
	},
	{
		Name:        "-",
		Types:       []string{"string"},
		Description: "Ignore/skip this field (use '-' as value)",
	},
}
