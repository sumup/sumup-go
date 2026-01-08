package builder

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"
)

func operationMethodName(op *openapi3.Operation) string {
	methodName := strcase.ToCamel(op.OperationID)
	if ext, ok := op.Extensions["x-codegen"]; ok {
		if extMap, ok := ext.(map[string]any); ok {
			if name, ok := extMap["method_name"]; ok {
				if nameString, ok := name.(string); ok {
					methodName = strcase.ToCamel(nameString)
				}
			}
		}
	}

	return methodName
}

func (b *Builder) schemaTypeName(ref string) string {
	if ref == "" {
		return ""
	}
	if override, ok := b.schemaNameOverrides[ref]; ok {
		return override
	}
	return strcase.ToCamel(strings.TrimPrefix(ref, "#/components/schemas/"))
}
