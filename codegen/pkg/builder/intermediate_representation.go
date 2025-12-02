package builder

import (
	"cmp"

	"github.com/getkin/kin-openapi/openapi3"
)

// TypeDeclaration holds the information for generating a type.
// TODO: split into struct, alias, etc.
type TypeDeclaration struct {
	// Name of the type
	Name string
	// Type describes the type of the type (e.g. struct, int64, string)
	Type string
	// Fields holds the information for the field
	Fields []StructField
	// Comment holds the description of the type
	Comment string

	// One of response, operation, or schema will be populated
	// based on what the type declaration was created from.

	Response  *openapi3.Response
	Operation *openapi3.Operation
	Schema    *openapi3.Schema
}

type OneOfDeclaration struct {
	Name    string
	Options []string
}

// StructField holds the information for StructField of a type.
type StructField struct {
	// Name of the field
	Name string
	// Type of the field, either primitive type (e.g. string) or if the field
	// is a schema reference then the type of the schema.
	Type string
	// Tags to apply to the field, this would usually be json serialization
	// information.
	Tags map[string][]string
	// Optional field.
	Optional bool
	// Pointer indicates whether the field should be a pointer in the generated struct.
	Pointer bool

	Comment string

	Parameter *openapi3.Parameter
}

type EnumOption[E cmp.Ordered] struct {
	Name  string
	Value E
}

// EnumDeclaration holds the information for enum types
type EnumDeclaration[E cmp.Ordered] struct {
	Type   TypeDeclaration
	Values []EnumOption[E]
}

type Response struct {
	IsErr          bool
	IsDefault      bool
	IsUnexpected   bool
	Type           string
	Code           int
	ErrDescription string
}
