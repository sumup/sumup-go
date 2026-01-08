package builder

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// operationGodoc creates godoc comment for an operation.
func operationGodoc(name string, operation *openapi3.Operation) string {
	out := new(strings.Builder)

	fmt.Fprintf(out, cmp.Or(operation.Description, operation.Summary, name))

	if operation.ExternalDocs != nil {
		extDescription := operation.ExternalDocs.Description
		if extDescription == "" {
			extDescription = "See"
		}

		fmt.Fprintf(out, "\n%s: %s", extDescription, operation.ExternalDocs.URL)
	}

	if operation.Deprecated {
		if notice, ok := operation.Extensions["x-deprecation-notice"].(string); ok {
			fmt.Fprintf(out, "\nDeprecated: %s", notice)
		} else {
			fmt.Fprint(out, "\nDeprecated: this operation is deprecated")
		}
	}

	return formatGodoc(out.String())
}

// operationParamsGodoc creates godoc comment for a struct representing
// parameters of an operation.
func operationParamsGodoc(name string, operation *openapi3.Operation) string {
	return formatGodoc(fmt.Sprintf("%s are query parameters for %s.", name, operation.OperationID))
}

// schemaGodoc creates godoc for a schema.
func schemaGodoc(name string, schema *openapi3.Schema) string {
	out := new(strings.Builder)

	if schema.Description != "" {
		fmt.Fprintf(out, schema.Description)
	} else {
		fmt.Fprintf(out, "%s is a schema definition.", name)
	}

	writeSchemaMetainfo(out, schema)

	return formatGodoc(out.String())
}

// schemaPropertyGodoc creates godoc for a schema property.
func schemaPropertyGodoc(s *openapi3.Schema) string {
	out := new(strings.Builder)

	fmt.Fprint(out, strings.TrimSpace(s.Description))

	writeSchemaMetainfo(out, s)

	return formatGodoc(out.String())
}

// parameterPropertyGodoc creates godoc for a request parameter property.
func parameterPropertyGodoc(s *openapi3.Parameter) string {
	out := new(strings.Builder)

	fmt.Fprint(out, strings.TrimSpace(s.Description))

	return formatGodoc(out.String())
}

// writeSchemaMetainfo writes additional schema metainfo such as validations
// into the output.
func writeSchemaMetainfo(out *strings.Builder, schema *openapi3.Schema) {
	if schema.ReadOnly {
		fmt.Fprintf(out, "\nRead only")
	}
	if schema.WriteOnly {
		fmt.Fprintf(out, "\nWrite only")
	}

	// add format but only if it can't be inferred from the type itself
	if !slices.Contains([]string{"", "date-time", "float"}, schema.Format) {
		fmt.Fprintf(out, "\nFormat: %v", schema.Format)
	}

	if schema.Default != nil {
		fmt.Fprintf(out, "\nDefault: %v", schema.Default)
	}

	// strings
	if schema.MinLength != 0 {
		fmt.Fprintf(out, "\nMin length: %v", schema.MinLength)
	}
	if schema.MaxLength != nil {
		fmt.Fprintf(out, "\nMax length: %v", *schema.MaxLength)
	}
	if schema.Pattern != "" {
		fmt.Fprintf(out, "\nPattern: %v", schema.Pattern)
	}

	// numbers
	if schema.Min != nil {
		fmt.Fprintf(out, "\nMin: %v", *schema.Min)
	}
	if schema.Max != nil {
		fmt.Fprintf(out, "\nMax: %v", *schema.Max)
	}
	if schema.MultipleOf != nil {
		fmt.Fprintf(out, "\nMultiple of: %v", *schema.MultipleOf)
	}

	// arrays
	if schema.UniqueItems {
		fmt.Fprintf(out, "\nUnique items only")
	}
	if schema.MinItems != 0 {
		fmt.Fprintf(out, "\nMin items: %v", schema.MinItems)
	}
	if schema.MaxItems != nil {
		fmt.Fprintf(out, "\nMax items: %v", *schema.MaxItems)
	}

	// objects
	if schema.MinProps != 0 {
		fmt.Fprintf(out, "\nMin properties: %v", schema.MinProps)
	}
	if schema.MaxProps != nil {
		fmt.Fprintf(out, "\nMax properties: %v", *schema.MaxProps)
	}

	if schema.ExternalDocs != nil {
		extDescription := schema.ExternalDocs.Description
		if extDescription == "" {
			extDescription = "See"
		}

		fmt.Fprintf(out, "\n%s: %s", extDescription, schema.ExternalDocs.URL)
	}

	if schema.Deprecated {
		if notice, ok := schema.Extensions["x-deprecation-notice"].(string); ok {
			fmt.Fprintf(out, "\nDeprecated: %s", notice)
		} else {
			fmt.Fprint(out, "\nDeprecated: this operation is deprecated")
		}
	}
}

// splitDocString inserts newlines into doc comments at approximately 100 character intervals.
func formatGodoc(s string) string {
	out := new(strings.Builder)

	sanitisedDescription := strings.ReplaceAll(s, "\n", "\n// ")

	var written int
	for _, subStr := range strings.SplitAfter(sanitisedDescription, " ") {
		if written > 100 {
			// Remove trailing space if inserting a newline.
			out.WriteString(strings.TrimSuffix(subStr, " "))
			out.WriteString("\n// ")
			written = 0

			continue
		}

		ct, _ := out.WriteString(subStr)
		written += ct

		if strings.Contains(subStr, "\n") {
			written = 0
		}
	}

	doc := strings.TrimSuffix(out.String(), "\n// ")
	doc = strings.TrimPrefix(doc, "\n// ")
	return doc
}
