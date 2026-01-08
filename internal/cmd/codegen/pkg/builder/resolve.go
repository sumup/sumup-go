package builder

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ResolveRef resolves a schema reference.
func ResolveRef(spec *openapi3.T, ref string) (*openapi3.SchemaRef, error) {
	// Split the reference string into its components
	components := strings.Split(ref, "/")
	if len(components) < 4 || components[0] != "#" || components[1] != "components" || components[2] != "schemas" {
		return nil, fmt.Errorf("invalid $ref: %s", ref)
	}

	switch components[1] {
	case "components":
		switch components[2] {
		case "schemas":
			refID := components[3]

			// Get the referenced schema object
			refSchema, ok := spec.Components.Schemas[refID]
			if !ok {
				return nil, fmt.Errorf("cannot resolve $ref: %s", ref)
			}
			return refSchema, nil
		default:
			return nil, fmt.Errorf("invalid component[1]: %q", components[2])
		}
	default:
		return nil, fmt.Errorf("invalid component[0]: %q", components[1])
	}
}
