package builder

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"
)

var (
	pathParamRegexp = regexp.MustCompile(`\{(\w+)\}`)
)

// Parameter is a method parameter.
type Parameter struct {
	Name string
	Type string
}

// Method describes a client method. Methods map one-to-one to OpenAPI operations.
type Method struct {
	Description  string
	HTTPMethod   string
	FunctionName string
	ResponseType *ResponseType
	Path         string
	PathParams   []Parameter
	QueryParams  *Parameter
	HasBody      bool
	Responses    []Response
}

func (mt Method) ParamsString() string {
	res := strings.Builder{}
	res.WriteString("ctx context.Context")
	for _, p := range mt.PathParams {
		res.WriteString(", ")
		res.WriteString(fmt.Sprintf("%s %s", strcase.ToLowerCamel(p.Name), p.Type))
	}
	if mt.QueryParams != nil {
		res.WriteString(", ")
		res.WriteString(fmt.Sprintf("%s %s", strcase.ToLowerCamel(mt.QueryParams.Name), mt.QueryParams.Type))
	}
	return res.String()
}

// pathsToMethods converts openapi3 path to golang methods.
func (b *Builder) pathsToMethods(tagName string, paths *v3.Paths) ([]*Method, error) {
	allMethods := make([]*Method, 0, paths.PathItems.Len())

	for _, path := range pathsInMatchingOrder(paths) {
		pathItem, ok := paths.PathItems.Get(path)
		if !ok {
			continue
		}
		if pathItem.IsReference() {
			slog.Warn(fmt.Sprintf("TODO: skipping path for %q, since it is a reference", path))
			continue
		}

		methods, err := b.pathToMethods(tagName, path, pathItem)
		if err != nil {
			return nil, err
		}

		allMethods = append(allMethods, methods...)
	}

	return allMethods, nil
}

// pathToMethods converts single openapi3 path to golang methods.
func (b *Builder) pathToMethods(tagName, path string, p *v3.PathItem) ([]*Method, error) {
	methods := make([]*Method, 0, p.GetOperations().Len())
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
		operationSpec.Parameters = append(operationSpec.Parameters, p.Parameters...)
		methodName := strings.ToUpper(method)
		method, err := b.operationToMethod(tagName, methodName, path, operationSpec)
		if err != nil {
			return nil, err
		}

		methods = append(methods, method)
	}

	return methods, nil
}

func pathBuilder(path string) string {
	var res strings.Builder

	params := make([]string, 0)
	for i, part := range strings.Split(path, "/") {
		if i != 0 {
			res.WriteString("/")
		}
		match := pathParamRegexp.FindStringSubmatch(part)
		if match == nil {
			res.WriteString(part)
		} else {
			res.WriteString("%v")
			params = append(params, strcase.ToLowerCamel(match[1]))
		}
	}

	if len(params) == 0 {
		return fmt.Sprintf("fmt.Sprintf(%q)", res.String())
	}

	return fmt.Sprintf("fmt.Sprintf(%q, %s)", res.String(), strings.Join(params, ", "))
}

func (b *Builder) operationToMethod(tagName, method, path string, o *v3.Operation) (*Method, error) {
	successResponses, err := b.collectSuccessResponses(o)
	if err != nil {
		return nil, fmt.Errorf("collect successful responses: %w", err)
	}
	singleSuccess := len(successResponses) == 1

	respType, err := b.getSuccessResponseType(tagName, o)
	if err != nil {
		return nil, fmt.Errorf("get successful response type: %w", err)
	}

	methodName := operationMethodName(o)
	typeName := b.operationTypeName(tagName, methodName)

	params, err := b.buildPathParams("path", o.Parameters)
	if err != nil {
		return nil, fmt.Errorf("build path parameters: %w", err)
	}

	hasBody := false
	if o.RequestBody != nil && o.RequestBody.Content != nil {
		mt, ok := getJSONMediaType(o.RequestBody.Content)
		if ok && mt.Schema != nil {
			paramType := typeName + "Params"
			params = append(params, Parameter{
				Name: "body",
				Type: paramType,
			})
			hasBody = true
		}
	}

	var queryParams *Parameter
	if slices.ContainsFunc(o.Parameters, func(p *v3.Parameter) bool {
		p = b.resolveParameter(p)
		return p != nil && p.In != "path"
	}) {
		queryParams = &Parameter{
			Name: "params",
			Type: typeName + "Params",
		}
	}

	responses := make([]Response, 0, o.Responses.Codes.Len())
	for code, resp := range o.Responses.Codes.FromOldest() {
		operationName := operationMethodName(o)
		typ := b.responseToType(tagName, operationName, resp, code, singleSuccess)

		description := code
		if resp.Description != "" {
			description = resp.Description
		}

		statusCode, _ := strconv.Atoi(strings.ReplaceAll(code, "XX", "00"))
		responses = append(responses, Response{
			IsErr:          !strings.HasPrefix(code, "2"),
			IsDefault:      code == "default",
			Type:           typ,
			Code:           statusCode,
			ErrDescription: strings.TrimSpace(description),
		})
	}

	slices.SortFunc(responses, func(a, b Response) int {
		if a.IsDefault {
			return 1000
		}

		if b.IsDefault {
			return -1000
		}

		return a.Code - b.Code
	})

	if !slices.ContainsFunc(responses, func(r Response) bool {
		return r.IsDefault
	}) {
		responses = append(responses, Response{
			IsErr:        false,
			IsDefault:    true,
			IsUnexpected: true,
			Type:         "",
			Code:         0,
		})
	}

	slog.Info("generating method",
		slog.String("id", o.OperationId),
		slog.String("method_name", methodName),
	)

	return &Method{
		Description:  operationGodoc(methodName, o),
		HTTPMethod:   httpMethod(method),
		FunctionName: methodName,
		ResponseType: respType,
		Path:         pathBuilder(path),
		PathParams:   params,
		QueryParams:  queryParams,
		HasBody:      hasBody,
		Responses:    responses,
	}, nil
}

type ResponseType struct {
	Type    string
	IsOneOf bool
}

func (b *Builder) getSuccessResponseType(tagName string, o *v3.Operation) (*ResponseType, error) {
	successResponses, err := b.collectSuccessResponses(o)
	if err != nil {
		return nil, err
	}

	if len(successResponses) == 0 {
		return nil, nil
	}

	if len(successResponses) == 1 {
		resp := successResponses[0]
		if resp.content.Schema != nil && resp.content.Schema.IsReference() {
			return &ResponseType{
				Type: b.getReferenceSchema(resp.content.Schema),
			}, nil
		}

		operationName := operationMethodName(o)
		typeName := b.operationTypeName(tagName, operationName)
		return &ResponseType{
			Type: b.getResponseName(typeName, resp.code, resp.content, true),
		}, nil
	}

	operationName := operationMethodName(o)
	typeName := b.operationTypeName(tagName, operationName)
	return &ResponseType{
		Type:    typeName + "Response",
		IsOneOf: true,
	}, nil
}

func (b *Builder) responseToType(tagName, operationName string, resp *v3.Response, code string, singleSuccess bool) string {
	if resp.IsReference() {
		return strcase.ToCamel(strings.TrimPrefix(resp.GetReference(), "#/components/responses/")) + "Response"
	}

	if resp.Content == nil {
		return ""
	}

	content, ok := getJSONMediaType(resp.Content)
	if !ok {
		return ""
	}

	if content.Schema == nil {
		return ""
	}

	if content.Schema.IsReference() {
		return b.getReferenceSchema(content.Schema)
	}

	if content.Schema.Schema() != nil {
		typeName := b.operationTypeName(tagName, operationName)
		isSuccess := strings.HasPrefix(code, "2")
		return b.getResponseName(typeName, code, content, isSuccess && singleSuccess)
	}

	return ""
}

type responseInfo struct {
	content *v3.MediaType
	code    string
}

func (b *Builder) collectSuccessResponses(o *v3.Operation) ([]responseInfo, error) {
	successResponses := make([]responseInfo, 0)
	for name, response := range o.Responses.Codes.FromOldest() {
		// TODO: throw error here?
		if name == "default" {
			name = "400"
		}

		statusCode, err := strconv.Atoi(strings.ReplaceAll(name, "XX", "00"))
		if err != nil {
			return nil, fmt.Errorf("error converting %q to an integer: %w", name, err)
		}

		if statusCode < 200 || statusCode >= 300 {
			// Continue early, we just want the successful response.
			continue
		}

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

		if content, ok := getJSONMediaType(response.Content); ok {
			if content.Schema != nil {
				successResponses = append(successResponses, responseInfo{
					content: content,
					code:    name,
				})
			}
		}
	}

	return successResponses, nil
}

func (b *Builder) buildPathParams(paramType string, params []*v3.Parameter) ([]Parameter, error) {
	if len(params) == 0 {
		return nil, nil
	}

	pathParams := make([]Parameter, 0)
	if paramType != "query" && paramType != "path" {
		return nil, errors.New("paramType must be one of 'query' or 'path'")
	}

	for _, p := range params {
		param := b.resolveParameter(p)
		if param == nil {
			if p != nil && p.Reference != "" {
				slog.Warn(fmt.Sprintf("param not resolved: %q", p.Reference))
			}
			continue
		}

		if param.In != "path" {
			continue
		}

		pathParams = append(pathParams, Parameter{
			Name: param.Name,
			Type: b.convertToValidGoType(param.Name, param.Schema),
		})
	}

	return pathParams, nil
}

// convertToValidGoType converts a schema type to a valid Go type.
func (b *Builder) convertToValidGoType(property string, r *base.SchemaProxy) string {
	if r == nil {
		return "any"
	}

	// Use reference as it is the type
	if r.IsReference() {
		return b.getReferenceSchema(r)
	}

	schema := r.Schema()
	if schema == nil {
		return "any"
	}

	if schema.AdditionalProperties != nil && schema.AdditionalProperties.IsA() && schema.AdditionalProperties.A != nil {
		propSchema := schema.AdditionalProperties.A
		if propSchema.IsReference() {
			return b.getReferenceSchema(propSchema)
		}
		if propSchema.Schema() != nil && propSchema.Schema().Items != nil && propSchema.Schema().Items.IsA() && propSchema.Schema().Items.A != nil {
			ref := b.getReferenceSchema(propSchema.Schema().Items.A)
			itemSchema := propSchema.Schema().Items.A.Schema()
			if itemSchema != nil && slices.Contains(itemSchema.Type, "array") {
				return "[]" + ref
			}
			return ref
		}
	}

	// TODO: Handle AllOf
	if schema.AllOf != nil {
		if len(schema.AllOf) > 1 {
			slog.Warn(fmt.Sprintf("TODO: allOf for %q has more than 1 item\n", property))
			return "TODO"
		}

		return b.convertToValidGoType(property, schema.AllOf[0])
	}

	switch {
	case slices.Contains(schema.Type, "string"):
		return formatStringType(schema)
	case slices.Contains(schema.Type, "integer"):
		return formatIntegerType(schema)
	case slices.Contains(schema.Type, "number"):
		return formatNumberType(schema)
	case slices.Contains(schema.Type, "boolean"):
		return "bool"
	case slices.Contains(schema.Type, "array"):
		if schema.Items != nil && schema.Items.IsA() && schema.Items.A != nil {
			reference := b.getReferenceSchema(schema.Items.A)
			if reference != "" {
				return fmt.Sprintf("[]%s", reference)
			}
		}
		// TODO: handle if it is not a reference.
		return "[]string"
	case slices.Contains(schema.Type, "object"):
		if schema.Properties == nil || schema.Properties.Len() == 0 {
			// TODO: generate type alias?
			slog.Warn("object with empty properties", slog.String("property", property))
			return "any"
		}
		// Most likely this is a local object, we will handle it.
		return strcase.ToCamel(property)
	default:
		slog.Warn("unknown type, falling back to 'any'",
			slog.String("property", property),
			slog.Any("type", schema.Type),
		)
		return "any"
	}
}

func (b *Builder) getReferenceSchema(v *base.SchemaProxy) string {
	if v == nil {
		return ""
	}

	if v.IsReference() {
		ref := strings.TrimPrefix(v.GetReference(), "#/components/schemas/")
		if v.Schema() != nil && len(v.Schema().Enum) > 0 {
			return strcase.ToCamel(strcase.MakeSingular(ref))
		}
		return strcase.ToCamel(ref)
	}

	return ""
}

// formatStringType converts a string schema to a valid Go type.
func formatStringType(t *base.Schema) string {
	if isNumericStringFormat(t.Format) {
		return "json.Number"
	}

	switch t.Format {
	case "date-time":
		return "time.Time"
	case "date":
		return "datetime.Date"
	case "time":
		return "datetime.Time"
	case "password":
		return "secret.Secret"
	default:
		return "string"
	}
}

func isNumericStringFormat(format string) bool {
	switch format {
	case "int", "int32", "int64", "integer", "float", "double", "decimal", "number":
		return true
	default:
		return false
	}
}

// formatIntegerType converts an integer schema to a valid Go type based on format.
func formatIntegerType(t *base.Schema) string {
	switch t.Format {
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	default:
		return "int"
	}
}

// formatNumberType converts a number schema to a valid Go type based on format.
func formatNumberType(t *base.Schema) string {
	switch t.Format {
	case "float":
		return "float32"
	case "double":
		return "float64"
	default:
		return "float64"
	}
}
