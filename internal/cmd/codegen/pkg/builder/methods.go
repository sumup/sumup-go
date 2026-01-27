package builder

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

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
func (b *Builder) pathsToMethods(tagName string, paths *openapi3.Paths) ([]*Method, error) {
	allMethods := make([]*Method, 0, paths.Len())

	for _, path := range paths.InMatchingOrder() {
		p := paths.Find(path)
		if p.Ref != "" {
			slog.Warn(fmt.Sprintf("TODO: skipping path for %q, since it is a reference", path))
			continue
		}

		methods, err := b.pathToMethods(tagName, path, p)
		if err != nil {
			return nil, err
		}

		allMethods = append(allMethods, methods...)
	}

	return allMethods, nil
}

// pathToMethods converts single openapi3 path to golang methods.
func (b *Builder) pathToMethods(tagName, path string, p *openapi3.PathItem) ([]*Method, error) {
	ops := p.Operations()
	keys := slices.Collect(maps.Keys(ops))
	slices.Sort(keys)

	methods := make([]*Method, 0, len(keys))
	for _, method := range keys {
		operationSpec := ops[method]
		operationSpec.Parameters = append(operationSpec.Parameters, p.Parameters...)
		method, err := b.operationToMethod(tagName, method, path, operationSpec)
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

func (b *Builder) operationToMethod(tagName, method, path string, o *openapi3.Operation) (*Method, error) {
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
	if o.RequestBody != nil {
		mt, ok := o.RequestBody.Value.Content["application/json"]
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
	if slices.ContainsFunc(o.Parameters, func(p *openapi3.ParameterRef) bool {
		return p.Value.In != "path"
	}) {
		queryParams = &Parameter{
			Name: "params",
			Type: typeName + "Params",
		}
	}

	responses := make([]Response, 0, o.Responses.Len())
	for code, resp := range o.Responses.Map() {
		operationName := operationMethodName(o)
		typ := b.responseToType(tagName, operationName, resp, code, singleSuccess)

		description := code
		if resp.Value.Description != nil {
			description = *resp.Value.Description
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
		slog.String("id", o.OperationID),
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

func (b *Builder) getSuccessResponseType(tagName string, o *openapi3.Operation) (*ResponseType, error) {
	successResponses, err := b.collectSuccessResponses(o)
	if err != nil {
		return nil, err
	}

	if len(successResponses) == 0 {
		return nil, nil
	}

	if len(successResponses) == 1 {
		resp := successResponses[0]
		if resp.content.Schema != nil && resp.content.Schema.Ref != "" {
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

func (b *Builder) responseToType(tagName, operationName string, resp *openapi3.ResponseRef, code string, singleSuccess bool) string {
	if resp.Ref != "" {
		return strcase.ToCamel(strings.TrimPrefix(resp.Ref, "#/components/responses/")) + "Response"
	}

	content, ok := resp.Value.Content["application/json"]
	if !ok {
		return ""
	}

	if content.Schema == nil {
		return ""
	}

	if content.Schema.Ref != "" {
		return b.getReferenceSchema(content.Schema)
	}

	if content.Schema.Value != nil {
		typeName := b.operationTypeName(tagName, operationName)
		isSuccess := strings.HasPrefix(code, "2")
		return b.getResponseName(typeName, code, content, isSuccess && singleSuccess)
	}

	return ""
}

type responseInfo struct {
	content *openapi3.MediaType
	code    string
}

func (b *Builder) collectSuccessResponses(o *openapi3.Operation) ([]responseInfo, error) {
	successResponses := make([]responseInfo, 0)
	for name, response := range o.Responses.Map() {
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

		if response.Ref != "" {
			ref := strings.TrimPrefix(response.Ref, "#/components/responses/")
			response = b.spec.Components.Responses[ref]
		}

		if content, ok := response.Value.Content["application/json"]; ok {
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

func (b *Builder) buildPathParams(paramType string, params openapi3.Parameters) ([]Parameter, error) {
	if len(params) == 0 {
		return nil, nil
	}

	pathParams := make([]Parameter, 0)
	if paramType != "query" && paramType != "path" {
		return nil, errors.New("paramType must be one of 'query' or 'path'")
	}

	for _, p := range params {
		if p.Value == nil {
			slog.Warn(fmt.Sprintf("param not resolved: %q", p.Ref))
			continue
		}

		if p.Value.In != "path" {
			continue
		}

		pathParams = append(pathParams, Parameter{
			Name: p.Value.Name,
			Type: b.convertToValidGoType(p.Value.Name, p.Value.Schema),
		})
	}

	return pathParams, nil
}

// convertToValidGoType converts a schema type to a valid Go type.
func (b *Builder) convertToValidGoType(property string, r *openapi3.SchemaRef) string {
	// Use reference as it is the type
	if r.Ref != "" {
		return b.getReferenceSchema(r)
	}

	if r.Value.AdditionalProperties.Schema != nil {
		if r.Value.AdditionalProperties.Schema.Ref != "" {
			return b.getReferenceSchema(r.Value.AdditionalProperties.Schema)
		} else if r.Value.AdditionalProperties.Schema.Value.Items.Ref != "" {
			ref := b.getReferenceSchema(r.Value.AdditionalProperties.Schema.Value.Items)
			if r.Value.AdditionalProperties.Schema.Value.Items.Value.Type.Is("array") {
				return "[]" + ref
			}
			return ref
		}
	}

	// TODO: Handle AllOf
	if r.Value.AllOf != nil {
		if len(r.Value.AllOf) > 1 {
			slog.Warn(fmt.Sprintf("TODO: allOf for %q has more than 1 item\n", property))
			return "TODO"
		}

		return b.convertToValidGoType(property, r.Value.AllOf[0])
	}

	switch {
	case r.Value.Type.Is("string"):
		return formatStringType(r.Value)
	case r.Value.Type.Is("integer"):
		return formatIntegerType(r.Value)
	case r.Value.Type.Is("number"):
		return formatNumberType(r.Value)
	case r.Value.Type.Is("boolean"):
		return "bool"
	case r.Value.Type.Is("array"):
		reference := b.getReferenceSchema(r.Value.Items)
		if reference != "" {
			return fmt.Sprintf("[]%s", reference)
		}
		// TODO: handle if it is not a reference.
		return "[]string"
	case r.Value.Type.Is("object"):
		if len(r.Value.Properties) == 0 {
			// TODO: generate type alias?
			slog.Warn("object with empty properties", slog.String("property", property))
			return "any"
		}
		// Most likely this is a local object, we will handle it.
		return strcase.ToCamel(property)
	default:
		slog.Warn("unknown type, falling back to 'any'",
			slog.Any("property", property),
			slog.Any("type", r.Value.Type),
		)
		return "any"
	}
}

func (b *Builder) getReferenceSchema(v *openapi3.SchemaRef) string {
	if v.Ref != "" {
		ref := strings.TrimPrefix(v.Ref, "#/components/schemas/")
		if len(v.Value.Enum) > 0 {
			return strcase.ToCamel(strcase.MakeSingular(ref))
		}
		return strcase.ToCamel(ref)
	}

	return ""
}

// formatStringType converts a string schema to a valid Go type.
func formatStringType(t *openapi3.Schema) string {
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

// formatIntegerType converts an integer schema to a valid Go type based on format.
func formatIntegerType(t *openapi3.Schema) string {
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
func formatNumberType(t *openapi3.Schema) string {
	switch t.Format {
	case "float":
		return "float32"
	case "double":
		return "float64"
	default:
		return "float64"
	}
}
