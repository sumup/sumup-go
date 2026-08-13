package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"go.yaml.in/yaml/v4"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"
)

const (
	sampleCatalogSchemaVersion = 1
	sdkModule                  = "github.com/sumup/sumup-go"
)

// SampleCatalog is the versioned JSON contract consumed by documentation sites.
type SampleCatalog struct {
	SchemaVersion  int      `json:"schemaVersion"`
	Language       string   `json:"language"`
	SDK            SDK      `json:"sdk"`
	OpenAPIVersion string   `json:"openAPIVersion"`
	Samples        []Sample `json:"samples"`
}

// SDK identifies the package used by every generated sample.
type SDK struct {
	Module  string `json:"module"`
	Version string `json:"version"`
}

// Sample is a complete, formatted Go program for one OpenAPI operation example.
type Sample struct {
	ID          string `json:"id"`
	OperationID string `json:"operationId"`
	Example     string `json:"example,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	HTTPMethod  string `json:"httpMethod"`
	Path        string `json:"path"`
	Source      string `json:"sample"`
}

// Samples builds a deterministic catalog of compilable Go examples.
func (b *Builder) Samples(sdkVersion string) (*SampleCatalog, error) {
	if b.spec == nil {
		return nil, fmt.Errorf("missing specs: call Load to load the specs first")
	}
	if b.spec.Info == nil {
		return nil, fmt.Errorf("missing specs info: call Load to load the specs first")
	}

	registry := b.sampleTypeRegistry()
	samples := make([]Sample, 0)
	for _, apiPath := range pathsInMatchingOrder(b.spec.Paths) {
		pathItem, ok := b.spec.Paths.PathItems.Get(apiPath)
		if !ok || pathItem == nil || pathItem.IsReference() {
			continue
		}

		operations := pathItem.GetOperations()
		methods := slices.Collect(operations.KeysFromOldest())
		slices.Sort(methods)
		for _, httpMethod := range methods {
			operation, ok := operations.Get(httpMethod)
			if !ok || operation == nil {
				continue
			}
			if operation.OperationId == "" {
				return nil, fmt.Errorf("missing operation id for %s %s", strings.ToUpper(httpMethod), apiPath)
			}

			operationCopy := *operation
			operationCopy.Parameters = append(slices.Clone(operation.Parameters), pathItem.Parameters...)
			tagName := "shared"
			if len(operation.Tags) > 0 {
				tagName = strings.ToLower(operation.Tags[0])
			}

			method, err := b.operationToMethod(tagName, strings.ToUpper(httpMethod), apiPath, &operationCopy)
			if err != nil {
				return nil, fmt.Errorf("build operation %q: %w", operation.OperationId, err)
			}

			operationSamples, err := b.samplesForOperation(
				registry,
				tagName,
				strings.ToUpper(httpMethod),
				apiPath,
				&operationCopy,
				method,
			)
			if err != nil {
				return nil, fmt.Errorf("generate samples for %q: %w", operation.OperationId, err)
			}
			samples = append(samples, operationSamples...)
		}
	}

	slices.SortFunc(samples, func(a, b Sample) int {
		return strings.Compare(a.ID, b.ID)
	})

	return &SampleCatalog{
		SchemaVersion: sampleCatalogSchemaVersion,
		Language:      "go",
		SDK: SDK{
			Module:  sdkModule,
			Version: sdkVersion,
		},
		OpenAPIVersion: strings.TrimSpace(b.spec.Info.Version),
		Samples:        samples,
	}, nil
}

type requestExample struct {
	name        string
	summary     string
	description string
	value       any
	provided    bool
}

func (b *Builder) samplesForOperation(
	registry sampleTypeRegistry,
	tagName string,
	httpMethod string,
	apiPath string,
	operation *v3.Operation,
	method *Method,
) ([]Sample, error) {
	examples := requestExamples(operation)
	samples := make([]Sample, 0, len(examples))
	for _, example := range examples {
		renderer := sampleRenderer{
			builder:  b,
			registry: registry,
			imports:  make(map[string]struct{}),
		}
		source, err := renderer.render(tagName, operation, method, example)
		if err != nil {
			return nil, err
		}

		id := operation.OperationId
		if example.name != "" {
			id += "." + example.name
		}
		summary := strings.TrimSpace(operation.Summary)
		if example.summary != "" {
			summary = strings.TrimSpace(example.summary)
		}
		description := strings.TrimSpace(operation.Description)
		if example.description != "" {
			description = strings.TrimSpace(example.description)
		}

		samples = append(samples, Sample{
			ID:          id,
			OperationID: operation.OperationId,
			Example:     example.name,
			Summary:     summary,
			Description: description,
			HTTPMethod:  httpMethod,
			Path:        apiPath,
			Source:      source,
		})
	}

	return samples, nil
}

func requestExamples(operation *v3.Operation) []requestExample {
	if operation.RequestBody == nil || operation.RequestBody.Content == nil {
		return []requestExample{{}}
	}
	mediaType, ok := getJSONMediaType(operation.RequestBody.Content)
	if !ok || mediaType == nil {
		return []requestExample{{}}
	}

	if mediaType.Examples != nil && mediaType.Examples.Len() > 0 {
		names := slices.Collect(mediaType.Examples.KeysFromOldest())
		slices.Sort(names)
		examples := make([]requestExample, 0, len(names))
		for _, name := range names {
			example, ok := mediaType.Examples.Get(name)
			if !ok || example == nil {
				continue
			}
			value, provided := decodeNode(example.Value)
			examples = append(examples, requestExample{
				name:        name,
				summary:     example.Summary,
				description: example.Description,
				value:       value,
				provided:    provided,
			})
		}
		if len(examples) > 0 {
			return examples
		}
	}

	if value, provided := decodeNode(mediaType.Example); provided {
		return []requestExample{{value: value, provided: true}}
	}
	if value, provided := schemaExample(mediaType.Schema); provided {
		return []requestExample{{value: value, provided: true}}
	}

	return []requestExample{{}}
}

type sampleTypeRegistry map[string]*TypeDeclaration

func (b *Builder) sampleTypeRegistry() sampleTypeRegistry {
	registry := make(sampleTypeRegistry)
	add := func(writables []Writable) {
		for _, writable := range writables {
			switch declaration := writable.(type) {
			case *TypeDeclaration:
				registry[declaration.Name] = declaration
			case *EnumDeclaration[string]:
				registry[declaration.Type.Name] = &declaration.Type
			case *EnumDeclaration[int]:
				registry[declaration.Type.Name] = &declaration.Type
			case *EnumDeclaration[int32]:
				registry[declaration.Type.Name] = &declaration.Type
			case *EnumDeclaration[int64]:
				registry[declaration.Type.Name] = &declaration.Type
			case *EnumDeclaration[float32]:
				registry[declaration.Type.Name] = &declaration.Type
			case *EnumDeclaration[float64]:
				registry[declaration.Type.Name] = &declaration.Type
			}
		}
	}

	schemaTags := make([]string, 0, len(b.resolvedSchemas))
	for tagName := range b.resolvedSchemas {
		schemaTags = append(schemaTags, tagName)
	}
	slices.Sort(schemaTags)
	for _, tagName := range schemaTags {
		add(b.schemasToTypes(b.resolvedSchemas[tagName], b.errorSchemas))
	}

	pathTags := make([]string, 0, len(b.pathsByTag))
	for tagName := range b.pathsByTag {
		pathTags = append(pathTags, tagName)
	}
	slices.Sort(pathTags)
	for _, tagName := range pathTags {
		add(b.pathsToBodyTypes(tagName, b.pathsByTag[tagName]))
		add(b.pathsToParamTypes(tagName, b.pathsByTag[tagName]))
	}

	return registry
}

type sampleRenderer struct {
	builder  *Builder
	registry sampleTypeRegistry
	imports  map[string]struct{}
}

func (r *sampleRenderer) render(
	tagName string,
	operation *v3.Operation,
	method *Method,
	example requestExample,
) (string, error) {
	r.imports["context"] = struct{}{}
	r.imports[sdkModule] = struct{}{}

	args := []string{"context.TODO()"}
	pathArgumentCount := len(method.PathParams)
	if method.HasBody {
		pathArgumentCount--
	}
	for i := range pathArgumentCount {
		argument := method.PathParams[i]
		parameter := r.parameter(operation.Parameters, argument.Name, "path")
		value, provided := parameterExample(parameter)
		expression, err := r.value(argument.Type, parameterSchema(parameter), value, provided)
		if err != nil {
			return "", fmt.Errorf("render path parameter %q: %w", argument.Name, err)
		}
		args = append(args, expression)
	}

	if method.HasBody {
		argument := method.PathParams[len(method.PathParams)-1]
		mediaType, _ := getJSONMediaType(operation.RequestBody.Content)
		expression, err := r.value(argument.Type, mediaType.Schema, example.value, example.provided)
		if err != nil {
			return "", fmt.Errorf("render request body: %w", err)
		}
		args = append(args, expression)
	}

	if method.QueryParams != nil {
		values := make(map[string]any)
		if declaration := r.registry.declaration(method.QueryParams.Type); declaration != nil {
			for _, field := range declaration.Fields {
				if value, provided := parameterExample(field.Parameter); provided {
					values[field.Name] = value
				}
			}
		}
		expression, err := r.value(method.QueryParams.Type, nil, values, false)
		if err != nil {
			return "", fmt.Errorf("render query parameters: %w", err)
		}
		args = append(args, expression)
	}

	call := fmt.Sprintf("client.%s.%s(%s)", strcase.ToCamel(tagName), method.FunctionName, strings.Join(args, ", "))
	var body strings.Builder
	body.WriteString("func main() {\n")
	body.WriteString("\tclient := sumup.NewClient()\n")
	if method.ResponseType == nil {
		fmt.Fprintf(&body, "\tif err := %s; err != nil {\n", call)
		body.WriteString("\t\tpanic(err.Error())\n")
		body.WriteString("\t}\n")
	} else {
		r.imports["fmt"] = struct{}{}
		fmt.Fprintf(&body, "\tresult, err := %s\n", call)
		body.WriteString("\tif err != nil {\n")
		body.WriteString("\t\tpanic(err.Error())\n")
		body.WriteString("\t}\n\n")
		body.WriteString("\tfmt.Printf(\"%+v\\n\", result)\n")
	}
	body.WriteString("}\n")

	var source bytes.Buffer
	source.WriteString("package main\n\n")
	source.WriteString(r.importBlock())
	source.WriteString("\n")
	source.WriteString(body.String())
	formatted, err := format.Source(source.Bytes())
	if err != nil {
		return "", fmt.Errorf("format source: %w\n%s", err, source.String())
	}
	return string(formatted), nil
}

func (r *sampleRenderer) parameter(parameters []*v3.Parameter, name, location string) *v3.Parameter {
	for _, parameter := range parameters {
		resolved := r.builder.resolveParameter(parameter)
		if resolved != nil && resolved.Name == name && resolved.In == location {
			return resolved
		}
	}
	return nil
}

func (r *sampleRenderer) value(
	typeName string,
	schema *base.SchemaProxy,
	raw any,
	provided bool,
) (string, error) {
	declaration := r.registry.declaration(typeName)
	if declaration != nil {
		switch {
		case declaration.Type == "struct":
			return r.structValue(typeName, declaration, raw, provided)
		case strings.HasPrefix(declaration.Type, "[]"):
			return r.sliceValue(typeName, strings.TrimPrefix(declaration.Type, "[]"), schema, raw, provided)
		case strings.HasPrefix(declaration.Type, "map["):
			return r.mapValue(typeName, raw)
		case declaration.Type == "json.RawMessage":
			return r.rawMessage(raw)
		case strings.HasPrefix(declaration.Type, "= "):
			aliased := strings.TrimSpace(strings.TrimPrefix(declaration.Type, "= "))
			target := r.registry.declaration(aliased)
			if target != nil && target.Type == "struct" {
				return r.structValue(typeName, target, raw, provided)
			}
			return r.primitiveValue(typeName, schema, raw, provided), nil
		default:
			return r.primitiveValue(typeName, schema, raw, provided), nil
		}
	}

	switch {
	case strings.HasPrefix(typeName, "[]"):
		return r.sliceValue(typeName, strings.TrimPrefix(typeName, "[]"), schema, raw, provided)
	case strings.HasPrefix(typeName, "map["):
		return r.mapValue(typeName, raw)
	case strings.HasPrefix(typeName, "nullable.Field["):
		inner := strings.TrimSuffix(strings.TrimPrefix(typeName, "nullable.Field["), "]")
		r.imports[sdkModule+"/nullable"] = struct{}{}
		if provided && raw == nil {
			return fmt.Sprintf("nullable.Null[%s]()", r.qualifyType(inner)), nil
		}
		expression, err := r.value(inner, schema, raw, provided)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("nullable.Value(%s)", expression), nil
	case typeName == "time.Time":
		r.imports["time"] = struct{}{}
		if text, ok := raw.(string); ok && text != "" {
			if timestamp, err := time.Parse(time.RFC3339, text); err == nil {
				location := "time.UTC"
				if _, offset := timestamp.Zone(); offset != 0 {
					location = fmt.Sprintf("time.FixedZone(%q, %d)", timestamp.Format("-07:00"), offset)
				}
				return fmt.Sprintf(
					"time.Date(%d, time.%s, %d, %d, %d, %d, %d, %s)",
					timestamp.Year(),
					timestamp.Month(),
					timestamp.Day(),
					timestamp.Hour(),
					timestamp.Minute(),
					timestamp.Second(),
					timestamp.Nanosecond(),
					location,
				), nil
			}
		}
		return "time.Time{}", nil
	case typeName == "datetime.Date":
		r.imports["time"] = struct{}{}
		r.imports[sdkModule+"/datetime"] = struct{}{}
		date := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
		if text, ok := raw.(string); ok {
			if parsed, err := time.Parse(time.DateOnly, text); err == nil {
				date = parsed
			}
		}
		return fmt.Sprintf("datetime.NewDate(%d, time.%s, %d)", date.Year(), date.Month(), date.Day()), nil
	case typeName == "datetime.Time":
		r.imports[sdkModule+"/datetime"] = struct{}{}
		clock := time.Date(0, time.January, 1, 12, 0, 0, 0, time.UTC)
		if text, ok := raw.(string); ok {
			if parsed, err := time.Parse(time.TimeOnly, text); err == nil {
				clock = parsed
			}
		}
		return fmt.Sprintf("datetime.NewTime(%d, %d, %d)", clock.Hour(), clock.Minute(), clock.Second()), nil
	case typeName == "secret.Secret":
		r.imports[sdkModule+"/secret"] = struct{}{}
		return fmt.Sprintf("secret.New(%q)", stringValue(raw, "secret")), nil
	case typeName == "json.Number":
		r.imports["encoding/json"] = struct{}{}
		return fmt.Sprintf("json.Number(%q)", stringValue(raw, "1")), nil
	case typeName == "any":
		return r.anyValue(raw), nil
	default:
		return r.primitiveValue(typeName, schema, raw, provided), nil
	}
}

func (r *sampleRenderer) structValue(
	typeName string,
	declaration *TypeDeclaration,
	raw any,
	provided bool,
) (string, error) {
	values, _ := raw.(map[string]any)
	var body strings.Builder
	fmt.Fprintf(&body, "%s{\n", r.qualifyType(typeName))
	for _, field := range declaration.Fields {
		if isReadOnly(field.Schema) {
			continue
		}
		key := field.Name
		value, fieldProvided := values[key]
		if !fieldProvided && (!provided || !field.Optional) {
			value, fieldProvided = schemaExample(field.Schema)
		}
		if field.Optional && !fieldProvided {
			continue
		}

		expression, err := r.value(field.Type, field.Schema, value, fieldProvided)
		if err != nil {
			return "", fmt.Errorf("render field %q: %w", field.Name, err)
		}
		if field.Pointer && !strings.HasPrefix(field.Type, "nullable.Field[") {
			expression = r.pointerValue(field.Type, expression)
		}
		fmt.Fprintf(&body, "%s: %s,\n", structFieldName(field.Name), expression)
	}
	body.WriteString("}")
	return body.String(), nil
}

func (r *sampleRenderer) pointerValue(typeName, expression string) string {
	if declaration := r.registry.declaration(typeName); declaration != nil && isPrimitiveDeclaration(declaration) {
		expression = fmt.Sprintf("%s(%s)", r.qualifyType(typeName), expression)
	} else {
		switch typeName {
		case "int32", "int64", "float32", "float64":
			expression = fmt.Sprintf("%s(%s)", typeName, expression)
		}
	}
	return fmt.Sprintf("new(%s)", expression)
}

func (r *sampleRenderer) sliceValue(
	typeName string,
	elementType string,
	schema *base.SchemaProxy,
	raw any,
	provided bool,
) (string, error) {
	items, _ := raw.([]any)
	var itemSchema *base.SchemaProxy
	if schema != nil && schema.Schema() != nil && schema.Schema().Items != nil && schema.Schema().Items.IsA() {
		itemSchema = schema.Schema().Items.A
	}
	var body strings.Builder
	fmt.Fprintf(&body, "%s{", r.qualifyType(typeName))
	for _, item := range items {
		expression, err := r.value(elementType, itemSchema, item, true)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&body, "%s,", expression)
	}
	body.WriteString("}")
	return body.String(), nil
}

func (r *sampleRenderer) mapValue(typeName string, raw any) (string, error) {
	values, _ := raw.(map[string]any)
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	var body strings.Builder
	fmt.Fprintf(&body, "%s{", r.qualifyType(typeName))
	for _, key := range keys {
		fmt.Fprintf(&body, "%q: %s,", key, r.anyValue(values[key]))
	}
	body.WriteString("}")
	return body.String(), nil
}

func (r *sampleRenderer) rawMessage(raw any) (string, error) {
	r.imports["encoding/json"] = struct{}{}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return "", fmt.Errorf("encode raw message: %w", err)
	}
	return fmt.Sprintf("json.RawMessage(%q)", string(encoded)), nil
}

func (r *sampleRenderer) primitiveValue(
	typeName string,
	schema *base.SchemaProxy,
	raw any,
	provided bool,
) string {
	if !provided {
		raw, provided = schemaExample(schema)
	}
	if !provided {
		raw = fallbackValue(typeName, schema)
	}

	switch value := raw.(type) {
	case string:
		return strconv.Quote(value)
	case bool:
		return strconv.FormatBool(value)
	case int:
		return strconv.Itoa(value)
	case int64:
		return strconv.FormatInt(value, 10)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case json.Number:
		return value.String()
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%v", value)
	}
}

func (r *sampleRenderer) anyValue(raw any) string {
	switch value := raw.(type) {
	case map[string]any:
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		var body strings.Builder
		body.WriteString("map[string]any{")
		for _, key := range keys {
			fmt.Fprintf(&body, "%q: %s,", key, r.anyValue(value[key]))
		}
		body.WriteString("}")
		return body.String()
	case []any:
		var body strings.Builder
		body.WriteString("[]any{")
		for _, item := range value {
			fmt.Fprintf(&body, "%s,", r.anyValue(item))
		}
		body.WriteString("}")
		return body.String()
	case string:
		return strconv.Quote(value)
	case bool:
		return strconv.FormatBool(value)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%v", value)
	}
}

func (r *sampleRenderer) qualifyType(typeName string) string {
	if strings.HasPrefix(typeName, "[]") {
		return "[]" + r.qualifyType(strings.TrimPrefix(typeName, "[]"))
	}
	if _, ok := r.registry[typeName]; ok {
		return "sumup." + typeName
	}
	return typeName
}

func (r *sampleRenderer) importBlock() string {
	standard := make([]string, 0)
	thirdParty := make([]string, 0)
	for path := range r.imports {
		if strings.Contains(path, ".") {
			thirdParty = append(thirdParty, path)
		} else {
			standard = append(standard, path)
		}
	}
	slices.Sort(standard)
	slices.Sort(thirdParty)

	var body strings.Builder
	body.WriteString("import (\n")
	for _, path := range standard {
		fmt.Fprintf(&body, "\t%q\n", path)
	}
	if len(standard) > 0 && len(thirdParty) > 0 {
		body.WriteString("\n")
	}
	for _, path := range thirdParty {
		fmt.Fprintf(&body, "\t%q\n", path)
	}
	body.WriteString(")\n")
	return body.String()
}

func (registry sampleTypeRegistry) declaration(typeName string) *TypeDeclaration {
	declaration := registry[typeName]
	seen := make(map[string]struct{})
	for declaration != nil && strings.HasPrefix(declaration.Type, "= ") {
		if _, ok := seen[declaration.Name]; ok {
			return declaration
		}
		seen[declaration.Name] = struct{}{}
		aliased := strings.TrimSpace(strings.TrimPrefix(declaration.Type, "= "))
		target := registry[aliased]
		if target == nil {
			return declaration
		}
		declaration = target
	}
	return declaration
}

func parameterSchema(parameter *v3.Parameter) *base.SchemaProxy {
	if parameter == nil {
		return nil
	}
	return parameter.Schema
}

func parameterExample(parameter *v3.Parameter) (any, bool) {
	if parameter == nil {
		return nil, false
	}
	if value, ok := decodeNode(parameter.Example); ok {
		return value, true
	}
	if parameter.Examples != nil {
		names := slices.Collect(parameter.Examples.KeysFromOldest())
		slices.Sort(names)
		for _, name := range names {
			example, ok := parameter.Examples.Get(name)
			if ok && example != nil {
				if value, ok := decodeNode(example.Value); ok {
					return value, true
				}
			}
		}
	}
	return schemaExample(parameter.Schema)
}

func schemaExample(schema *base.SchemaProxy) (any, bool) {
	if schema == nil || schema.Schema() == nil {
		return nil, false
	}
	spec := schema.Schema()
	if value, ok := decodeNode(spec.Example); ok {
		return value, true
	}
	for _, example := range spec.Examples {
		if value, ok := decodeNode(example); ok {
			return value, true
		}
	}
	if value, ok := decodeNode(spec.Default); ok {
		return value, true
	}
	if len(spec.Enum) > 0 {
		return decodeNode(spec.Enum[0])
	}
	return nil, false
}

func decodeNode(node *yaml.Node) (any, bool) {
	if node == nil {
		return nil, false
	}
	var value any
	if err := node.Decode(&value); err != nil {
		return nil, false
	}
	return value, true
}

func fallbackValue(typeName string, schema *base.SchemaProxy) any {
	if schema != nil && schema.Schema() != nil {
		spec := schema.Schema()
		switch {
		case slices.Contains(spec.Type, "string"):
			switch spec.Format {
			case "date-time":
				return "2025-01-01T00:00:00Z"
			case "date":
				return "2025-01-01"
			case "time":
				return "12:00:00"
			case "email":
				return "developer@example.com"
			case "uri", "url":
				return "https://example.com"
			case "uuid":
				return "00000000-0000-4000-8000-000000000000"
			default:
				return "string"
			}
		case slices.Contains(spec.Type, "integer"):
			return float64(1)
		case slices.Contains(spec.Type, "number"):
			return 1.0
		case slices.Contains(spec.Type, "boolean"):
			return true
		}
	}

	switch typeName {
	case "string", "secret.Secret", "json.Number":
		return "string"
	case "bool":
		return true
	case "int", "int32", "int64", "float32", "float64":
		return float64(1)
	default:
		return "string"
	}
}

func stringValue(value any, fallback string) string {
	if value == nil {
		return fallback
	}
	return fmt.Sprint(value)
}

func isReadOnly(schema *base.SchemaProxy) bool {
	return schema != nil && schema.Schema() != nil && schema.Schema().ReadOnly != nil && *schema.Schema().ReadOnly
}

func isPrimitiveDeclaration(declaration *TypeDeclaration) bool {
	if declaration == nil {
		return false
	}
	return declaration.Type != "struct" &&
		!strings.HasPrefix(declaration.Type, "[]") &&
		!strings.HasPrefix(declaration.Type, "map[") &&
		declaration.Type != "json.RawMessage"
}
