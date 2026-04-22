package builder

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"
)

type OperationSample struct {
	Sample string `json:"sample"`
}

func (b *Builder) BuildSamples() (map[string]map[string]OperationSample, error) {
	if b.spec == nil {
		return nil, fmt.Errorf("missing specs: call Load to load the specs first")
	}

	samples := make(map[string]map[string]OperationSample)
	for _, path := range pathsInMatchingOrder(b.spec.Paths) {
		pathItem, ok := b.spec.Paths.PathItems.Get(path)
		if !ok || pathItem == nil || pathItem.IsReference() {
			continue
		}

		methodSamples, err := b.pathToSamples(path, pathItem)
		if err != nil {
			return nil, err
		}
		if len(methodSamples) == 0 {
			continue
		}

		samples[path] = methodSamples
	}

	return samples, nil
}

func (b *Builder) pathToSamples(path string, p *v3.PathItem) (map[string]OperationSample, error) {
	samples := make(map[string]OperationSample)

	ops := p.GetOperations()
	opKeys := make([]string, 0, ops.Len())
	for key := range ops.KeysFromOldest() {
		opKeys = append(opKeys, key)
	}
	slices.Sort(opKeys)

	for _, method := range opKeys {
		operationSpec, ok := ops.Get(method)
		if !ok {
			continue
		}

		sample, err := b.operationSample(p.Parameters, operationSpec)
		if err != nil {
			return nil, fmt.Errorf("build sample for %s %s: %w", strings.ToUpper(method), path, err)
		}

		samples[strings.ToUpper(method)] = OperationSample{Sample: sample}
	}

	return samples, nil
}

func (b *Builder) operationSample(pathParams []*v3.Parameter, op *v3.Operation) (string, error) {
	tagName := "shared"
	switch {
	case len(op.Tags) > 1:
		tagName = strings.ToLower(op.Tags[0])
	case len(op.Tags) == 1:
		tagName = strings.ToLower(op.Tags[0])
	}

	service := strcase.ToCamel(b.tagByTagName(tagName).Name)
	methodName := operationMethodName(op)
	typeName := b.operationTypeName(tagName, methodName) + "Params"

	params := append(slices.Clone(pathParams), op.Parameters...)
	callArgs := []string{"context.Background()"}

	for _, p := range params {
		param := b.resolveParameter(p)
		if param == nil || param.In != "path" {
			continue
		}
		callArgs = append(callArgs, b.sampleExprForParameter(param))
	}

	if body := operationJSONRequestBody(op); body != nil && body.Schema != nil {
		bodySample, err := b.sampleExprForSchema(sampleTypeInfo{
			Expr: "sumup." + typeName,
			Bare: typeName,
		}, body.Schema, sampleValueFromMediaType(body), 1, true)
		if err != nil {
			return "", err
		}
		callArgs = append(callArgs, bodySample)
	}

	if queryParams := nonPathParameters(b, params); len(queryParams) > 0 {
		querySample, err := b.sampleExprForQueryParams(typeName, queryParams, 1)
		if err != nil {
			return "", err
		}
		callArgs = append(callArgs, querySample)
	}

	assign := "result, err := "
	respType, err := b.getSuccessResponseType(tagName, op)
	if err != nil {
		return "", err
	}
	if respType == nil {
		assign = "err := "
	}

	var out strings.Builder
	out.WriteString("client := sumup.NewClient()\n\n")
	out.WriteString(assign)
	out.WriteString("client.")
	out.WriteString(service)
	out.WriteString(".")
	out.WriteString(methodName)
	out.WriteString("(\n")
	for _, arg := range callArgs {
		out.WriteString(indentSample(arg, 1))
		out.WriteString(",\n")
	}
	out.WriteString(")")

	return out.String(), nil
}

func nonPathParameters(b *Builder, params []*v3.Parameter) []*v3.Parameter {
	var queryParams []*v3.Parameter
	for _, p := range params {
		param := b.resolveParameter(p)
		if param == nil || param.In == "path" {
			continue
		}
		queryParams = append(queryParams, param)
	}
	return queryParams
}

func operationJSONRequestBody(op *v3.Operation) *v3.MediaType {
	if op == nil || op.RequestBody == nil || op.RequestBody.Content == nil {
		return nil
	}

	mt, ok := getJSONMediaType(op.RequestBody.Content)
	if !ok {
		return nil
	}

	return mt
}

func (b *Builder) sampleExprForQueryParams(typeName string, params []*v3.Parameter, level int) (string, error) {
	lines := make([]string, 0, len(params))
	for _, param := range params {
		if param == nil {
			continue
		}

		example, hasExample := sampleValueFromParameter(param)
		if param.Required == nil || !*param.Required {
			continue
		}
		if !hasExample && !schemaHasExample(param.Schema) {
			example = nil
		}

		fieldValue, err := b.sampleExprForSchema(sampleTypeInfo{}, param.Schema, example, level+1, false)
		if err != nil {
			return "", err
		}
		lines = append(lines, fmt.Sprintf("%s: %s,", sampleFieldName(param.Name), fieldValue))
	}

	if len(lines) == 0 && len(params) > 0 {
		param := params[0]
		fieldValue, err := b.sampleExprForSchema(sampleTypeInfo{}, param.Schema, nil, level+1, false)
		if err != nil {
			return "", err
		}
		lines = append(lines, fmt.Sprintf("%s: %s,", sampleFieldName(param.Name), fieldValue))
	}

	return renderCompositeLiteral("sumup."+typeName, lines, level), nil
}

type sampleTypeInfo struct {
	Expr string
	Bare string
}

func (b *Builder) sampleExprForParameter(param *v3.Parameter) string {
	if param == nil {
		return `""`
	}

	example, _ := sampleValueFromParameter(param)
	expr, err := b.sampleExprForSchema(sampleTypeInfo{}, param.Schema, example, 0, false)
	if err != nil {
		return `""`
	}
	return expr
}

func (b *Builder) sampleExprForSchema(typeInfo sampleTypeInfo, schema *base.SchemaProxy, example any, level int, topLevel bool) (string, error) {
	if schema == nil {
		return "nil", nil
	}

	if schema.IsReference() {
		resolved, err := b.resolveSchemaRef(schema.GetReference())
		if err != nil {
			return "", err
		}

		refName := b.getReferenceSchema(schema)
		if typeInfo.Bare == "" {
			typeInfo.Bare = refName
		}
		if typeInfo.Expr == "" {
			typeInfo.Expr = "sumup." + refName
		}

		return b.sampleExprForSchema(typeInfo, resolved, example, level, topLevel)
	}

	spec := schema.Schema()
	if spec == nil {
		return "nil", nil
	}

	if example == nil {
		example, _ = sampleValueFromSchema(spec)
	}

	switch {
	case len(spec.Enum) > 0:
		if example == nil {
			example = sampleValueFromEnum(spec.Enum)
		}
	case spec.Const != nil && example == nil:
		example, _ = sampleValueFromYAMLNode(spec.Const)
	}

	switch {
	case isAdditionalPropertiesMap(spec):
		if m, ok := example.(map[string]any); ok && len(m) != 0 {
			return renderMapLiteral(m, level), nil
		}
		return `map[string]any{"key": "value"}`, nil
	case isFreeFormObject(spec):
		if example != nil {
			raw, err := json.Marshal(example)
			if err == nil {
				return fmt.Sprintf("json.RawMessage(%q)", string(raw)), nil
			}
		}
		return `json.RawMessage(` + "`" + `{"key":"value"}` + "`" + `)`, nil
	case slices.Contains(spec.Type, "object") || spec.AllOf != nil:
		exampleMap, _ := example.(map[string]any)
		fields := make([]string, 0)
		properties := orderedKeys(spec.Properties)
		for property := range spec.Properties.FromOldest() {
			propSchema, ok := spec.Properties.Get(property)
			if !ok {
				continue
			}

			propValue, hasPropValue := exampleMap[property]
			include := slices.Contains(spec.Required, property)
			if !include {
				continue
			}
			if !hasPropValue {
				propValue = nil
			}

			fieldType := b.sampleFieldType(typeInfo.Bare, property, propSchema)
			fieldExpr, err := b.sampleExprForSchema(fieldType, propSchema, propValue, level+1, false)
			if err != nil {
				return "", err
			}
			fields = append(fields, fmt.Sprintf("%s: %s,", sampleFieldName(property), fieldExpr))
		}

		if len(fields) == 0 && len(properties) > 0 {
			property := properties[0]
			propSchema, _ := spec.Properties.Get(property)
			fieldType := b.sampleFieldType(typeInfo.Bare, property, propSchema)
			fieldExpr, err := b.sampleExprForSchema(fieldType, propSchema, nil, level+1, false)
			if err != nil {
				return "", err
			}
			fields = append(fields, fmt.Sprintf("%s: %s,", sampleFieldName(property), fieldExpr))
		}

		return renderCompositeLiteral(typeInfo.Expr, fields, level), nil
	case slices.Contains(spec.Type, "array"):
		itemExample := any(nil)
		if examples, ok := example.([]any); ok && len(examples) > 0 {
			itemExample = examples[0]
		}

		itemSchema := spec.Items
		if itemSchema == nil || !itemSchema.IsA() || itemSchema.A == nil {
			return "nil", nil
		}

		itemType := b.sampleArrayItemType(typeInfo.Bare, spec.Items.A)
		itemExpr, err := b.sampleExprForSchema(itemType, itemSchema.A, itemExample, level+1, false)
		if err != nil {
			return "", err
		}

		if typeInfo.Expr != "" {
			return renderCompositeLiteral(typeInfo.Expr, []string{itemExpr + ","}, level), nil
		}
		itemTypeExpr := "[]any"
		if itemType.Expr != "" {
			itemTypeExpr = "[]" + strings.TrimPrefix(itemType.Expr, "[]")
		}
		return renderCompositeLiteral(itemTypeExpr, []string{itemExpr + ","}, level), nil
	case slices.Contains(spec.Type, "string"):
		return sampleStringExpr(spec, example, typeInfo.Expr, topLevel), nil
	case slices.Contains(spec.Type, "integer"):
		return sampleIntegerExpr(spec, example, typeInfo.Expr, topLevel), nil
	case slices.Contains(spec.Type, "number"):
		return sampleNumberExpr(spec, example, typeInfo.Expr, topLevel), nil
	case slices.Contains(spec.Type, "boolean"):
		if example == nil {
			example = true
		}
		return renderTypedScalar(typeInfo.Expr, fmt.Sprintf("%t", example), topLevel), nil
	default:
		return "nil", nil
	}
}

func (b *Builder) sampleFieldType(parentBare, property string, schema *base.SchemaProxy) sampleTypeInfo {
	if schema == nil {
		return sampleTypeInfo{}
	}

	if schema.IsReference() {
		refName := b.getReferenceSchema(schema)
		return sampleTypeInfo{
			Expr: "sumup." + refName,
			Bare: refName,
		}
	}

	spec := schema.Schema()
	if spec == nil {
		return sampleTypeInfo{}
	}

	if slices.Contains(spec.Type, "object") || spec.AllOf != nil {
		typeName := parentBare + strcase.ToCamel(property)
		return sampleTypeInfo{
			Expr: "sumup." + typeName,
			Bare: typeName,
		}
	}
	if slices.Contains(spec.Type, "array") && spec.Items != nil && spec.Items.IsA() && spec.Items.A != nil {
		item := spec.Items.A
		if item.IsReference() {
			refName := b.getReferenceSchema(item)
			return sampleTypeInfo{
				Expr: "[]" + "sumup." + refName,
				Bare: refName,
			}
		}
		if item.Schema() != nil && (slices.Contains(item.Schema().Type, "object") || item.Schema().AllOf != nil) {
			typeName := strcase.MakeSingular(parentBare + strcase.ToCamel(property))
			return sampleTypeInfo{
				Expr: "[]" + "sumup." + typeName,
				Bare: typeName,
			}
		}
		return sampleTypeInfo{
			Expr: "[]" + b.convertToValidGoType(property, item),
		}
	}

	return sampleTypeInfo{}
}

func (b *Builder) sampleArrayItemType(parentBare string, schema *base.SchemaProxy) sampleTypeInfo {
	if schema == nil {
		return sampleTypeInfo{}
	}

	if schema.IsReference() {
		refName := b.getReferenceSchema(schema)
		return sampleTypeInfo{
			Expr: "sumup." + refName,
			Bare: refName,
		}
	}

	spec := schema.Schema()
	if spec == nil {
		return sampleTypeInfo{}
	}

	if slices.Contains(spec.Type, "object") || spec.AllOf != nil {
		typeName := strcase.MakeSingular(parentBare)
		return sampleTypeInfo{
			Expr: "sumup." + typeName,
			Bare: typeName,
		}
	}

	return sampleTypeInfo{}
}

func sampleStringExpr(schema *base.Schema, example any, typeExpr string, topLevel bool) string {
	if example == nil {
		example = sampleDefaultString(schema)
	}

	value, _ := example.(string)
	switch {
	case isNumericStringFormat(schema.Format):
		return renderTypedScalar(typeExpr, fmt.Sprintf("json.Number(%q)", value), topLevel)
	case schema.Format == "password":
		return fmt.Sprintf("secret.New(%q)", value)
	case schema.Format == "date-time":
		return "time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)"
	case schema.Format == "date":
		return "datetime.NewDate(2025, time.January, 1)"
	case schema.Format == "time":
		return "datetime.NewTime(12, 0, 0)"
	default:
		return renderTypedScalar(typeExpr, strconv.Quote(value), topLevel)
	}
}

func sampleIntegerExpr(_ *base.Schema, example any, typeExpr string, topLevel bool) string {
	value := "1"
	switch v := example.(type) {
	case int:
		value = strconv.Itoa(v)
	case int64:
		value = strconv.FormatInt(v, 10)
	case float64:
		value = strconv.FormatInt(int64(v), 10)
	}
	return renderTypedScalar(typeExpr, value, topLevel)
}

func sampleNumberExpr(_ *base.Schema, example any, typeExpr string, topLevel bool) string {
	value := "10.1"
	switch v := example.(type) {
	case float32:
		value = strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		value = strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		value = strconv.Itoa(v)
	case int64:
		value = strconv.FormatInt(v, 10)
	}
	return renderTypedScalar(typeExpr, value, topLevel)
}

func sampleDefaultString(schema *base.Schema) string {
	switch schema.Format {
	case "uuid":
		return "00000000-0000-0000-0000-000000000000"
	case "uri", "url":
		return "https://example.com"
	case "email":
		return "user@example.com"
	case "date":
		return "2025-01-01"
	case "time":
		return "12:00:00"
	case "date-time":
		return time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	case "password":
		return "secret"
	default:
		return "example"
	}
}

func renderTypedScalar(typeExpr, value string, topLevel bool) string {
	if topLevel && typeExpr != "" {
		return fmt.Sprintf("%s(%s)", typeExpr, value)
	}
	return value
}

func renderCompositeLiteral(typeExpr string, lines []string, level int) string {
	if len(lines) == 0 {
		if typeExpr == "" {
			return "{}"
		}
		return typeExpr + "{}"
	}

	var out strings.Builder
	if typeExpr != "" {
		out.WriteString(typeExpr)
	}
	out.WriteString("{\n")
	for _, line := range lines {
		out.WriteString(indentSample(line, level+1))
		out.WriteString("\n")
	}
	out.WriteString(strings.Repeat("\t", level))
	out.WriteString("}")
	return out.String()
}

func renderMapLiteral(values map[string]any, level int) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		value := sampleAnyExpr(values[key])
		lines = append(lines, fmt.Sprintf("%q: %s,", key, value))
	}
	return renderCompositeLiteral("map[string]any", lines, level)
}

func sampleAnyExpr(value any) string {
	switch v := value.(type) {
	case string:
		return strconv.Quote(v)
	case bool:
		return fmt.Sprintf("%t", v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case []any:
		values := make([]string, 0, len(v))
		for _, item := range v {
			values = append(values, sampleAnyExpr(item)+",")
		}
		return renderCompositeLiteral("[]any", values, 0)
	case map[string]any:
		return renderMapLiteral(v, 0)
	default:
		return `"value"`
	}
}

func indentSample(s string, level int) string {
	prefix := strings.Repeat("\t", level)
	return prefix + strings.ReplaceAll(s, "\n", "\n"+prefix)
}

func sampleFieldName(name string) string {
	switch {
	case strings.HasPrefix(name, "+"):
		name = strings.Replace(name, "+", "Plus", 1)
	case strings.HasPrefix(name, "-"):
		name = strings.Replace(name, "-", "Minus", 1)
	case strings.HasPrefix(name, "@"):
		name = strings.Replace(name, "@", "At", 1)
	case strings.HasPrefix(name, "$"):
		name = strings.Replace(name, "$", "", 1)
	}
	return strcase.ToCamel(name)
}

func orderedKeys(properties *orderedmap.Map[string, *base.SchemaProxy]) []string {
	if properties == nil {
		return nil
	}

	keys := make([]string, 0, properties.Len())
	for key := range properties.FromOldest() {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func sampleValueFromParameter(param *v3.Parameter) (any, bool) {
	if param == nil {
		return nil, false
	}
	if value, ok := sampleValueFromYAMLNode(param.Example); ok {
		return value, true
	}
	if param.Examples != nil {
		for _, example := range param.Examples.FromOldest() {
			if example == nil {
				continue
			}
			if value, ok := sampleValueFromYAMLNode(example.Value); ok {
				return value, true
			}
		}
	}
	if param.Schema != nil {
		return sampleValueFromSchemaProxy(param.Schema)
	}
	return nil, false
}

func sampleValueFromMediaType(mt *v3.MediaType) any {
	if mt == nil {
		return nil
	}
	if value, ok := sampleValueFromYAMLNode(mt.Example); ok {
		return value
	}
	if mt.Examples != nil {
		for _, example := range mt.Examples.FromOldest() {
			if example == nil {
				continue
			}
			if value, ok := sampleValueFromYAMLNode(example.Value); ok {
				return value
			}
		}
	}
	if mt.Schema != nil {
		value, _ := sampleValueFromSchemaProxy(mt.Schema)
		return value
	}
	return nil
}

func sampleValueFromSchemaProxy(schema *base.SchemaProxy) (any, bool) {
	if schema == nil || schema.Schema() == nil {
		return nil, false
	}
	return sampleValueFromSchema(schema.Schema())
}

func sampleValueFromSchema(schema *base.Schema) (any, bool) {
	if schema == nil {
		return nil, false
	}
	if value, ok := sampleValueFromYAMLNode(schema.Example); ok {
		return value, true
	}
	for _, example := range schema.Examples {
		if value, ok := sampleValueFromYAMLNode(example); ok {
			return value, true
		}
	}
	if value, ok := sampleValueFromYAMLNode(schema.Default); ok {
		return value, true
	}
	return nil, false
}

func sampleValueFromEnum(enum []*yaml.Node) any {
	for _, value := range enum {
		if decoded, ok := sampleValueFromYAMLNode(value); ok {
			return decoded
		}
	}
	return nil
}

func sampleValueFromYAMLNode(node *yaml.Node) (any, bool) {
	if node == nil {
		return nil, false
	}

	var value any
	if err := node.Decode(&value); err != nil {
		return nil, false
	}
	return value, true
}

func schemaHasExample(schema *base.SchemaProxy) bool {
	if schema == nil {
		return false
	}
	if schema.IsReference() {
		return true
	}
	spec := schema.Schema()
	if spec == nil {
		return false
	}
	return spec.Example != nil || len(spec.Examples) > 0 || spec.Default != nil || spec.Const != nil || len(spec.Enum) > 0
}
