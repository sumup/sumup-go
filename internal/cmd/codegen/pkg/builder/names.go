package builder

import (
	"strings"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"
	"github.com/sumup/sumup-go/internal/cmd/codegen/pkg/extension"
)

func operationMethodName(op *v3.Operation) string {
	methodName := strcase.ToCamel(op.OperationId)
	if ext, ok := extension.Get[map[string]any](op.Extensions, "x-codegen"); ok {
		if name, ok := ext["method_name"]; ok {
			if nameString, ok := name.(string); ok {
				methodName = strcase.ToCamel(nameString)
			}
		}
	}

	return methodName
}

func (b *Builder) operationTypePrefix(tagName string) string {
	if tagName == "" || tagName == "shared" {
		return ""
	}
	return strcase.ToCamel(tagName)
}

func (b *Builder) operationTypeName(tagName, base string) string {
	return b.operationTypePrefix(tagName) + base
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
