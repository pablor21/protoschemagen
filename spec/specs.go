package spec

import "github.com/pablor21/gonnotation/annotations"

type AnnotationSpec = annotations.AnnotationSpec
type ValidOn = annotations.AnnotationValidOn
type Param = annotations.AnnotationParam
type TagParam = annotations.TagParam

// Specs aggregates Protobuf annotation and struct tag specs for the format generator
var Specs = annotations.PluginDefinitions{
	Annotations: annotationSpecs,
	StructTags:  tagParams,
}
