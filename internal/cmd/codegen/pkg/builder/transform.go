package builder

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"
)

// schemasToTypes converts openapi3 schemas to golang struct and enum types.
func (b *Builder) schemasToTypes(schemas []NamedSchema, errorSchemas map[string]struct{}) []Writable {
	var allTypes []Writable

	for _, s := range schemas {
		ref := s.Ref
		_, isErr := errorSchemas[ref]
		name := b.schemaTypeName(ref)
		typeTpl := b.generateSchemaComponents(name, s.Schema, isErr)
		allTypes = append(allTypes, typeTpl...)
	}

	return allTypes
}

// schemasToTypes converts openapi3 schemas to golang struct and enum types.
func (b *Builder) respToTypes(schemas []*v3.Response, errorSchemas map[string]struct{}) []Writable {
	var allTypes []Writable

	for _, s := range schemas {
		_, isErr := errorSchemas[s.Reference]
		name := strcase.ToCamel(strings.TrimPrefix(s.Reference, "#/components/responses/")) + "Response"
		if s.Content == nil {
			if isErr {
				allTypes = append(allTypes, typeAssertionDeclaration{
					typ: name,
				})
			}

			description := s.Description

			allTypes = append(allTypes, &TypeDeclaration{
				Comment:  description,
				Type:     "struct{}",
				Name:     name,
				Response: s,
			})
			if isErr {
				allTypes = append(allTypes, staticErrorImplementation{
					Typ:  name,
					Name: description,
				})
			}
			continue
		}

		content, ok := getJSONMediaType(s.Content)
		if !ok || content.Schema == nil {
			continue
		}

		typeTpl := b.generateSchemaComponents(name, content.Schema, isErr)
		allTypes = append(allTypes, typeTpl...)
	}

	return allTypes
}

// TODO: is this different from respToTypes?
func (b *Builder) pathsToBodyTypes(tagName string, paths *v3.Paths) []Writable {
	if paths == nil {
		return nil
	}

	paramTypes := make([]Writable, 0)
	aliasTypes := make(map[string]bool)
	for _, path := range pathsInMatchingOrder(paths) {
		pathSpec, ok := paths.PathItems.Get(path)
		if !ok {
			continue
		}
		if pathSpec.IsReference() {
			slog.Warn(fmt.Sprintf("TODO: skipping path for %q, since it is a reference", path))
			continue
		}

		for _, opSpec := range pathSpec.GetOperations().FromOldest() {
			operationName := operationMethodName(opSpec)
			typeName := b.operationTypeName(tagName, operationName)

			if opSpec.RequestBody != nil && opSpec.RequestBody.Content != nil {
				mt, ok := getJSONMediaType(opSpec.RequestBody.Content)
				if ok && mt.Schema != nil {
					paramsName := typeName + "Params"
					if mt.Schema.IsReference() {
						baseName := b.schemaTypeName(mt.Schema.GetReference())
						if paramsName != baseName && !aliasTypes[paramsName] {
							paramTypes = append(paramTypes, &TypeDeclaration{
								Name: paramsName,
								Type: "= " + baseName,
							})
							aliasTypes[paramsName] = true
						}
						continue
					}
					if mt.Schema.Schema() == nil {
						continue
					}
					bodyObject, additionalTypes := b.createObject(mt.Schema.Schema(), paramsName)
					paramTypes = append(paramTypes, bodyObject)
					paramTypes = append(paramTypes, additionalTypes...)
				}
			}
		}
	}

	return paramTypes
}

// constructParamTypes constructs struct for query parameters for an operation.
func (b *Builder) pathsToParamTypes(tagName string, paths *v3.Paths) []Writable {
	if paths == nil {
		return nil
	}

	paramTypes := make([]Writable, 0)

	for _, path := range pathsInMatchingOrder(paths) {
		pathSpec, ok := paths.PathItems.Get(path)
		if !ok {
			continue
		}
		if pathSpec.IsReference() {
			slog.Warn(fmt.Sprintf("TODO: skipping path for %q, since it is a reference", path))
			continue
		}

		for _, opSpec := range pathSpec.GetOperations().FromOldest() {
			operationName := operationMethodName(opSpec)
			typeName := b.operationTypeName(tagName, operationName)

			if len(opSpec.Parameters) > 0 {
				fields := make([]StructField, 0)
				for _, p := range opSpec.Parameters {
					param := b.resolveParameter(p)
					if param == nil {
						if p != nil && p.Reference != "" {
							slog.Warn(fmt.Sprintf("param not resolved: %q", p.Reference))
						}
						continue
					}

					// path parameters are passed as a parameters to the generated method
					if param.In == "path" {
						continue
					}

					name := param.Name

					typ := b.convertToValidGoType("", param.Schema)

					optional := param.Required == nil || !*param.Required
					pointer := shouldUsePointer(optional, param.Schema, typ)
					fields = append(fields, StructField{
						Name:      name,
						Type:      typ,
						Parameter: param,
						Optional:  optional,
						Pointer:   pointer,
						Comment:   parameterPropertyGodoc(param),
					})
				}

				if len(fields) != 0 {
					paramsTypeName := typeName + "Params"
					paramsTpl := TypeDeclaration{
						Type:      "struct",
						Name:      paramsTypeName,
						Comment:   operationParamsGodoc(paramsTypeName, opSpec),
						Fields:    fields,
						Operation: opSpec,
					}

					paramTypes = append(paramTypes, &paramsTpl, &toQueryValues{Typ: &paramsTpl})
				}
			}
		}
	}

	return paramTypes
}

// pathsToResponseTypes generates response types for operations. This is responsible only for inlined
// schemas that are specific to the operation itself and are not references.
func (b *Builder) pathsToResponseTypes(tagName string, paths *v3.Paths) []Writable {
	if paths == nil {
		return nil
	}

	paramTypes := make([]Writable, 0)

	for _, path := range pathsInMatchingOrder(paths) {
		pathSpec, ok := paths.PathItems.Get(path)
		if !ok {
			continue
		}
		if pathSpec.IsReference() {
			slog.Warn(fmt.Sprintf("TODO: skipping path for %q, since it is a reference", path))
			continue
		}

		for _, opSpec := range pathSpec.GetOperations().FromOldest() {
			operationName := operationMethodName(opSpec)
			typeName := b.operationTypeName(tagName, operationName)

			if opSpec.Responses == nil {
				continue
			}

			responses := opSpec.Responses.Codes
			responseKeys := make([]string, 0, responses.Len())
			for key := range responses.KeysFromOldest() {
				responseKeys = append(responseKeys, key)
			}

			slices.Sort(responseKeys)

			successInfos, err := b.collectSuccessResponses(opSpec)
			if err != nil {
				slog.Warn("failed to collect success responses",
					slog.Any("error", err),
					slog.String("operation_id", opSpec.OperationId),
				)
			}
			singleSuccess := len(successInfos) == 1

			var successResponses []string
			for _, code := range responseKeys {
				response, ok := responses.Get(code)
				if !ok {
					continue
				}
				isSuccess := strings.HasPrefix(code, "2")
				isErr := code == "default" || strings.HasPrefix(code, "4") || strings.HasPrefix(code, "5")

				if response.Reference != "" {
					resolved := b.resolveResponseRef(response.Reference)
					if resolved == nil {
						continue
					}
					response = resolved
				}

				if response.Content == nil {
					continue
				}

				content, ok := getJSONMediaType(response.Content)
				if !ok {
					continue
				}

				if content.Schema == nil {
					continue
				}

				if content.Schema.IsReference() {
					if isSuccess {
						name := strcase.ToCamel(strings.TrimPrefix(content.Schema.GetReference(), "#/components/schemas/"))
						successResponses = append(successResponses, name)
					}
					// schemas are handled separately, here we only care about inline schemas in the operation
					continue
				}

				name := b.getResponseName(typeName, code, content, isSuccess && singleSuccess)

				objects := b.generateSchemaComponents(name, content.Schema, isErr)
				paramTypes = append(paramTypes, objects...)

				if strings.HasPrefix(code, "2") {
					if resp, ok := objects[0].(*TypeDeclaration); ok {
						successResponses = append(successResponses, resp.Name)
					}
				}
			}

			// if there are multiple success responses, we need to create a oneOf type
			if len(successResponses) > 1 {
				slog.Info("multiple success responses found",
					slog.Any("responses", successResponses),
				)

				paramTypes = append(paramTypes, &OneOfDeclaration{
					Name:    typeName + "Response",
					Options: successResponses,
				})
			}
		}
	}

	return paramTypes
}

// generateSchemaComponents generates types from schema reference.
// This should be used to generate top-level types, that is - named schemas that are listed
// in `#/components/schemas/` part of the OpenAPI specs.
func (b *Builder) generateSchemaComponents(name string, schema *base.SchemaProxy, isErr bool) []Writable {
	types := make([]Writable, 0)
	if schema == nil || schema.Schema() == nil {
		return types
	}

	spec := schema.Schema()

	switch {
	case len(spec.Enum) > 0:
		enum := createEnum(spec, name)
		if enum != nil {
			types = append(types, enum)
		}
	case slices.Contains(spec.Type, "string"):
		types = append(types, &TypeDeclaration{
			Comment: schemaGodoc(name, spec),
			Type:    "string",
			Name:    name,
			Schema:  spec,
		})
	case slices.Contains(spec.Type, "integer"):
		types = append(types, &TypeDeclaration{
			Comment: schemaGodoc(name, spec),
			Type:    formatIntegerType(spec),
			Name:    name,
			Schema:  spec,
		})
	case slices.Contains(spec.Type, "number"):
		types = append(types, &TypeDeclaration{
			Comment: schemaGodoc(name, spec),
			Type:    formatNumberType(spec),
			Name:    name,
			Schema:  spec,
		})
	case slices.Contains(spec.Type, "boolean"):
		types = append(types, &TypeDeclaration{
			Comment: schemaGodoc(name, spec),
			Type:    "bool",
			Name:    name,
			Schema:  spec,
		})
	case slices.Contains(spec.Type, "array"):
		if spec.Items == nil || !spec.Items.IsA() || spec.Items.A == nil {
			return types
		}
		typeName, itemTypes := b.genSchema(spec.Items.A, strcase.MakeSingular(name))
		types = append(types, itemTypes...)
		types = append(types, &TypeDeclaration{
			Comment: schemaGodoc(name, spec),
			Type:    fmt.Sprintf("[]%s", typeName),
			Name:    name,
			Schema:  spec,
		})
	case slices.Contains(spec.Type, "object"):
		object, additionalTypes := b.createObject(spec, name)
		types = append(types, object)
		types = append(types, additionalTypes...)

		if isErr {
			types = append(types, errorImplementation{
				Typ: object,
			})
		}
	case spec.OneOf != nil:
		object := createOneOf(spec, name)
		types = append(types, object)
		if isErr {
			types = append(types, errorImplementation{
				Typ: object,
			})
		}
	case spec.AnyOf != nil:
		slog.Warn("AnyOf not supported, falling back to 'any'",
			slog.Any("name", name),
		)
		types = append(types, &TypeDeclaration{
			Comment: schemaGodoc(name, spec),
			Type:    "any",
			Name:    name,
			Schema:  spec,
		})
	case spec.AllOf != nil:
		object, additionalTypes := b.createAllOf(spec, name)
		types = append(types, object)
		types = append(types, additionalTypes...)
		if isErr {
			types = append(types, errorImplementation{
				Typ: object,
			})
		}
	default:
		if spec.Type != nil {
			slog.Warn("skipping unknown type",
				slog.Any("name", name),
				slog.Any("type", spec.Type),
			)
		}
	}

	if isErr {
		types = append(types, typeAssertionDeclaration{
			typ: name,
		})
	}

	return types
}

// genSchema is very similar to [generateSchemaComponents] but assumes that all schema components
// have been already generated.
func (b *Builder) genSchema(schema *base.SchemaProxy, name string) (string, []Writable) {
	if schema == nil {
		return "any", nil
	}

	if schema.IsReference() {
		ref := strings.TrimPrefix(schema.GetReference(), "#/components/schemas/")
		if schema.Schema() != nil && len(schema.Schema().Enum) > 0 {
			return strcase.ToCamel(strcase.MakeSingular(ref)), nil
		}

		return strcase.ToCamel(ref), nil
	}

	types := make([]Writable, 0)
	spec := schema.Schema()
	if spec == nil {
		return "any", nil
	}

	switch {
	case len(spec.Enum) > 0:
		enum := createEnum(spec, name)
		if enum != nil {
			types = append(types, enum)
		}
		return strcase.MakeSingular(name), types
	case slices.Contains(spec.Type, "string"):
		return formatStringType(spec), nil
	case slices.Contains(spec.Type, "integer"):
		return formatIntegerType(spec), nil
	case slices.Contains(spec.Type, "number"):
		return formatNumberType(spec), nil
	case slices.Contains(spec.Type, "boolean"):
		return "bool", nil
	case slices.Contains(spec.Type, "array"):
		if spec.Items == nil || !spec.Items.IsA() || spec.Items.A == nil {
			return "[]any", nil
		}
		typeName, schemas := b.genSchema(spec.Items.A, strcase.MakeSingular(name))
		types = append(types, schemas...)
		return "[]" + typeName, types
	case slices.Contains(spec.Type, "object"):
		object, additionalTypes := b.createObject(spec, name)
		types = append(types, object)
		types = append(types, additionalTypes...)
		return name, types
	case spec.OneOf != nil:
		object := createOneOf(spec, name)
		types = append(types, object)
		return name, types
	case spec.AnyOf != nil:
		slog.Warn("AnyOf not supported, falling back to 'any'",
			slog.Any("name", name),
		)
		return "any", nil
	case spec.AllOf != nil:
		object, additionalTypes := b.createAllOf(spec, name)
		types = append(types, object)
		types = append(types, additionalTypes...)
		return name, types
	default:
		if spec.Type != nil {
			slog.Warn("skipping unknown type",
				slog.Any("name", name),
				slog.Any("type", spec.Type),
			)
		}
		return "any", nil
	}
}

func isAdditionalPropertiesMap(schema *base.Schema) bool {
	if schema == nil {
		return false
	}
	if schema.Properties != nil && schema.Properties.Len() != 0 {
		return false
	}
	if schema.AdditionalProperties == nil {
		return false
	}
	if schema.AdditionalProperties.IsB() {
		return schema.AdditionalProperties.B
	}
	return schema.AdditionalProperties.IsA() && schema.AdditionalProperties.A != nil
}

func isArraySchema(schema *base.Schema) bool {
	if schema == nil {
		return false
	}
	return slices.Contains(schema.Type, "array")
}

func shouldUsePointer(optional bool, schema *base.SchemaProxy, typeName string) bool {
	if !optional {
		return false
	}
	if schema != nil && schema.Schema() != nil {
		if isAdditionalPropertiesMap(schema.Schema()) || isArraySchema(schema.Schema()) {
			return false
		}
	}
	if strings.HasPrefix(typeName, "[]") {
		return false
	}
	return true
}

// createObject converts openapi schema into golang object.
func (b *Builder) createObject(schema *base.Schema, name string) (*TypeDeclaration, []Writable) {
	if isAdditionalPropertiesMap(schema) {
		return &TypeDeclaration{
			Comment: schemaGodoc(name, schema),
			Name:    name,
			Type:    "map[string]any",
			Schema:  schema,
		}, nil
	}

	fields, additionalTypes := b.createFields(schema.Properties, name, schema.Required)
	return &TypeDeclaration{
		Comment: schemaGodoc(name, schema),
		Name:    name,
		Type:    "struct",
		Fields:  fields,
		Schema:  schema,
	}, additionalTypes
}

// createFields returns list of fields for openapi schema properties.
func (b *Builder) createFields(properties *orderedmap.Map[string, *base.SchemaProxy], name string, required []string) ([]StructField, []Writable) {
	fields := []StructField{}
	types := []Writable{}
	if properties == nil {
		return fields, types
	}

	keys := make([]string, 0, properties.Len())
	for key := range properties.FromOldest() {
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return fields, types
	}

	slices.Sort(keys)

	for _, property := range keys {
		schema, ok := properties.Get(property)
		if !ok {
			continue
		}
		typeName, moreTypes := b.genSchema(schema, name+strcase.ToCamel(property))

		tags := []string{strcase.ToSnake(property)}
		if !slices.Contains(required, property) {
			tags = append(tags, "omitempty")
		}
		optional := !slices.Contains(required, property)
		pointer := shouldUsePointer(optional, schema, typeName)
		fields = append(fields, StructField{
			Name:    property,
			Type:    typeName,
			Comment: schemaPropertyGodoc(schema.Schema()),
			Tags: map[string][]string{
				"json": tags,
			},
			Optional: optional,
			Pointer:  pointer,
		})
		types = append(types, moreTypes...)
	}

	return fields, types
}

func createEnum(schema *base.Schema, name string) Writable {
	enumName := strcase.MakeSingular(name)
	switch {
	case slices.Contains(schema.Type, "string"):
		values := make([]EnumOption[string], 0)
		for _, v := range schema.Enum {
			var option string
			if err := v.Decode(&option); err != nil {
				slog.Warn("invalid enum value",
					slog.String("enum", name),
					slog.String("expected", "string"),
					slog.String("got", fmt.Sprintf("%T", v)),
					slog.String("err", err.Error()),
					slog.Any("raw", v),
				)
				continue
			}

			values = append(values, EnumOption[string]{
				Name:  enumName + strcase.ToCamel(option),
				Value: option,
			})
		}

		return &EnumDeclaration[string]{
			Type: TypeDeclaration{
				Comment: schemaGodoc(name, schema),
				Name:    enumName,
				Type:    "string",
				Schema:  schema,
			},
			Values: values,
		}
	case slices.Contains(schema.Type, "integer"):
		if schema.Format == "int64" {
			values := make([]EnumOption[int64], 0)
			for _, v := range schema.Enum {
				var option float64
				if err := v.Decode(&option); err != nil {
					slog.Warn("invalid enum value",
						slog.String("enum", name),
						slog.String("expected", "int64"),
						slog.String("got", fmt.Sprintf("%T", v)),
						slog.String("err", err.Error()),
						slog.Any("raw", v),
					)
					continue
				}

				values = append(values, EnumOption[int64]{
					Name:  enumName + strcase.ToCamel(fmt.Sprintf("%v", option)),
					Value: int64(option),
				})
			}

			return &EnumDeclaration[int64]{
				Type: TypeDeclaration{
					Comment: schemaGodoc(name, schema),
					Name:    strcase.MakeSingular(name),
					Type:    "int64",
					Schema:  schema,
				},
				Values: values,
			}
		}

		if schema.Format == "int32" {
			values := make([]EnumOption[int32], 0)
			for _, v := range schema.Enum {
				var option float64
				if err := v.Decode(&option); err != nil {
					slog.Warn("invalid enum value",
						slog.String("enum", name),
						slog.String("expected", "int32"),
						slog.String("got", fmt.Sprintf("%T", v)),
						slog.String("err", err.Error()),
						slog.Any("raw", v),
					)
					continue
				}

				values = append(values, EnumOption[int32]{
					Name:  enumName + strcase.ToCamel(fmt.Sprintf("%v", option)),
					Value: int32(option),
				})
			}

			return &EnumDeclaration[int32]{
				Type: TypeDeclaration{
					Comment: schemaGodoc(name, schema),
					Name:    strcase.MakeSingular(name),
					Type:    "int32",
					Schema:  schema,
				},
				Values: values,
			}
		}

		values := make([]EnumOption[int], 0)
		for _, v := range schema.Enum {
			var option float64
			if err := v.Decode(&option); err != nil {
				slog.Warn("invalid enum value",
					slog.String("enum", name),
					slog.String("expected", "int"),
					slog.String("got", fmt.Sprintf("%T", v)),
					slog.String("err", err.Error()),
					slog.Any("raw", v),
				)
				continue
			}

			values = append(values, EnumOption[int]{
				Name:  enumName + strcase.ToCamel(fmt.Sprintf("%v", option)),
				Value: int(option),
			})
		}

		return &EnumDeclaration[int]{
			Type: TypeDeclaration{
				Comment: schemaGodoc(name, schema),
				Name:    strcase.MakeSingular(name),
				Type:    "int",
				Schema:  schema,
			},
			Values: values,
		}
	case slices.Contains(schema.Type, "number"):
		if schema.Format == "float" {
			values := make([]EnumOption[float32], 0)
			for _, v := range schema.Enum {
				var option float64
				if err := v.Decode(&option); err != nil {
					slog.Warn("invalid enum value",
						slog.String("enum", name),
						slog.String("expected", "float32"),
						slog.String("got", fmt.Sprintf("%T", v)),
						slog.String("err", err.Error()),
						slog.Any("raw", v),
					)
					continue
				}

				values = append(values, EnumOption[float32]{
					Name:  enumName + strcase.ToCamel(fmt.Sprintf("%v", option)),
					Value: float32(option),
				})
			}

			return &EnumDeclaration[float32]{
				Type: TypeDeclaration{
					Comment: schemaGodoc(name, schema),
					Name:    strcase.MakeSingular(name),
					Type:    "float32",
					Schema:  schema,
				},
				Values: values,
			}
		}

		values := make([]EnumOption[float64], 0)
		for _, v := range schema.Enum {
			var option float64
			if err := v.Decode(&option); err != nil {
				slog.Warn("invalid enum value",
					slog.String("enum", name),
					slog.String("expected", "float64"),
					slog.String("got", fmt.Sprintf("%T", v)),
					slog.String("err", err.Error()),
					slog.Any("raw", v),
				)
				continue
			}

			values = append(values, EnumOption[float64]{
				Name:  enumName + strcase.ToCamel(fmt.Sprintf("%v", option)),
				Value: option,
			})
		}

		return &EnumDeclaration[float64]{
			Type: TypeDeclaration{
				Comment: schemaGodoc(name, schema),
				Name:    strcase.MakeSingular(name),
				Type:    "float64",
				Schema:  schema,
			},
			Values: values,
		}
	default:
		return nil
	}
}

// createAllOf creates a type declaration for `allOf` schema.
func (b *Builder) createAllOf(schema *base.Schema, name string) (*TypeDeclaration, []Writable) {
	types := []Writable{}
	var fields []StructField
	seen := make(map[string]struct{})

	for _, s := range schema.AllOf {
		if s == nil || s.Schema() == nil {
			continue
		}

		properties := s.Schema().Properties
		if properties == nil {
			continue
		}

		filtered := orderedmap.New[string, *base.SchemaProxy]()
		for propName, propSchema := range properties.FromOldest() {
			if _, ok := seen[propName]; ok {
				continue
			}
			filtered.Set(propName, propSchema)
			seen[propName] = struct{}{}
		}

		objectFields, additionalTypes := b.createFields(filtered, name, s.Schema().Required)
		fields = append(fields, objectFields...)
		types = append(types, additionalTypes...)
	}

	return &TypeDeclaration{
		Comment: schemaGodoc(name, schema),
		Name:    name,
		Type:    "struct",
		Fields:  uniqueFields(fields),
		Schema:  schema,
	}, types
}

// createOneOf creates a type declaration for `oneOf` schema.
func createOneOf(schema *base.Schema, name string) *TypeDeclaration {
	// TODO: implement `func (v *{{name}}) AsXXX() (XXX, error) { ... }`
	// that allows converting one of from `json.RawMessage` to possible variants.

	return &TypeDeclaration{
		Comment: schemaGodoc(name, schema),
		Name:    name,
		Type:    "json.RawMessage",
		Schema:  schema,
	}
}

func uniqueFields(fields []StructField) []StructField {
	return uniqueFunc(fields, func(f StructField) string { return f.Name })
}

func uniqueFunc[T any, C comparable](arr []T, keyFn func(T) C) []T {
	seen := make(map[C]bool)

	n := 0
	for _, ele := range arr {
		key := keyFn(ele)
		if ok := seen[key]; ok {
			continue
		}
		arr[n] = ele
		n++
		seen[key] = true
	}

	return arr[:n]
}

func (b *Builder) getResponseName(operationName, responseCode string, content *v3.MediaType, singleSuccess bool) string {
	if singleSuccess {
		return operationName + "Response"
	}

	if content.Schema != nil && content.Schema.Schema() != nil && content.Schema.Schema().Title != "" {
		return operationName + strcase.ToCamel(content.Schema.Schema().Title) + "Response"
	}

	return operationName + responseCode + "Response"
}
