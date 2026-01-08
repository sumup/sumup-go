package builder

import (
	"fmt"
	"log/slog"
	"maps"
	"path"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"
	"github.com/sumup/sumup-go/internal/cmd/codegen/templates"
)

// Builder is an SDK builder. Builder works in two steps:
// First, it loads the OpenAPI specs and pre-processes them for code generation
// by transforming the specs into intermediary representation.
// Secondly, it generates and writes the SDK to desired destination.
type Builder struct {
	cfg Config

	// spec are the OpenAPI specs we are working with.
	// Load the specs using [Builder.Load].
	spec *openapi3.T

	// schemasByTag maps tags to respective schema references.
	schemasByTag    map[string][]string
	resolvedSchemas map[string][]*openapi3.SchemaRef

	responsesByTag    map[string][]string
	resolvedResponses map[string][]*openapi3.ResponseRef

	// errorSchemas are refs of schemas that are used for error responses (status code >= 400).
	errorSchemas map[string]struct{}
	// schemaNameOverrides maps schema refs to overridden type names.
	schemaNameOverrides map[string]string
	pathsByTag          map[string]*openapi3.Paths

	templates *template.Template

	start time.Time
}

// Config is builder configuration which configures output options.
type Config struct {
	// Out is the output directory that the SDK will be written to.
	Out string
	// Name is the name of the product / service.
	Name string
}

type Option func(b *Builder)

// New creates a new [Builder]. Call [Build.Load] to load in OpenAPI specs and
// [Build.Build] to generate SDK based on provided config.
func New(cfg Config, opts ...Option) *Builder {
	templates, err := template.New("").Funcs(template.FuncMap{
		"httpMethod":     httpMethod,
		"httpStatusCode": httpStatusCode,
	}).ParseFS(templates.Templates, "*")
	if err != nil {
		panic(err)
	}

	b := &Builder{
		cfg:                 cfg,
		schemasByTag:        make(map[string][]string),
		resolvedSchemas:     make(map[string][]*openapi3.SchemaRef),
		resolvedResponses:   make(map[string][]*openapi3.ResponseRef),
		pathsByTag:          make(map[string]*openapi3.Paths),
		errorSchemas:        make(map[string]struct{}),
		schemaNameOverrides: make(map[string]string),
		templates:           templates,
	}

	for _, o := range opts {
		o(b)
	}

	return b
}

// Load loads OpenAPI specs into the builder.
// To generated the SDK, call [Builder.Build].
func (b *Builder) Load(spec *openapi3.T) error {
	b.start = time.Now()
	b.spec = spec

	b.collectPaths()

	b.collectSchemas()
	if err := b.resolveSchemas(); err != nil {
		return fmt.Errorf("resolve schemas: %w", err)
	}

	b.collectResponses()
	b.resolveResponses()

	return nil
}

// Build the SDK and write it to designated output directory.
// The OpenAPI specs first need to be loaded using [Builder.Load].
func (b *Builder) Build() error {
	if b.spec == nil {
		return fmt.Errorf("missing specs: call Load to load the specs first")
	}

	if _, ok := b.pathsByTag["shared"]; !ok {
		b.pathsByTag["shared"] = openapi3.NewPaths()
	}

	for tagName, paths := range b.pathsByTag {
		if err := b.generateResource(tagName, paths); err != nil {
			return err
		}
	}

	if err := b.writeClientPackage(path.Join(b.cfg.Out, "client/client.go")); err != nil {
		return err
	}

	if err := b.writeClientFile(path.Join(b.cfg.Out, "client.go"), slices.Collect(maps.Keys(b.pathsByTag))); err != nil {
		return err
	}

	took := time.Since(b.start)
	slog.Info("sdk generated", slog.Duration("took", took))

	return nil
}

func (b *Builder) collectPaths() {
	for path, pathSpec := range b.spec.Paths.Map() {
		for method, operationSpec := range pathSpec.Operations() {
			var tag string
			if len(operationSpec.Tags) > 1 {
				slog.Warn("multiple tags for operation, picking first tag",
					slog.String("path", path),
					slog.String("method", method),
				)

				tag = strings.ToLower(operationSpec.Tags[0])
			}

			if len(operationSpec.Tags) == 1 {
				tag = strings.ToLower(operationSpec.Tags[0])
			}

			if len(operationSpec.Tags) == 0 {
				slog.Warn("no tags for operation, classifying as 'shared'",
					slog.String("path", path),
					slog.String("method", method),
				)
				tag = "shared"
			}

			tagPaths, ok := b.pathsByTag[tag]
			if !ok {
				tagPaths = openapi3.NewPaths()
				b.pathsByTag[tag] = tagPaths
			}

			tagPaths.Set(path, pathSpec)
		}
	}
}

func (b *Builder) resolveResponses() {
	for tag, refs := range b.responsesByTag {
		for _, ref := range refs {
			name := strings.TrimPrefix(ref, "#/components/responses/")
			resolved := b.spec.Components.Responses[name]
			resolved.Ref = ref
			b.resolvedResponses[tag] = append(b.resolvedResponses[tag], resolved)
		}
	}
}

func (b *Builder) resolveSchemas() error {
	for tag, refs := range b.schemasByTag {
		for _, ref := range refs {
			resolved, err := ResolveRef(b.spec, ref)
			if err != nil {
				return fmt.Errorf("resolve reference %q: %w", ref, err)
			}

			// TODO: remove, currently used for struct name in [generateTypes].
			resolved.Ref = ref
			b.resolvedSchemas[tag] = append(b.resolvedSchemas[tag], resolved)
		}
	}

	return nil
}

func (b *Builder) collectSchemas() {
	// Map of schemas grouped by tag
	schemasByTag := make(map[string][]string)

	// Map of schemas to list of tags that reference them
	schemaRefs := make(map[string][]string)

	for _, pathItem := range b.spec.Paths.Map() {
		for _, op := range pathItem.Operations() {
			operationName := strcase.ToCamel(op.OperationID)
			methodName := operationMethodName(op)

			requestSchemas := collectSchemasInRequest(op)
			schemas := b.collectSchemasInResponse(op)
			schemas = append(schemas, collectSchemasInParams(op)...)
			schemas = append(schemas, requestSchemas...)
			if methodName != operationName {
				for _, schema := range requestSchemas {
					if schema == nil || schema.Ref == "" {
						continue
					}

					refName := strings.TrimPrefix(schema.Ref, "#/components/schemas/")
					if !strings.HasPrefix(refName, operationName) {
						continue
					}

					suffix := strings.TrimPrefix(refName, operationName)
					if suffix == "" {
						continue
					}

					b.schemaNameOverrides[schema.Ref] = methodName + suffix
				}
			}

			for _, schema := range schemas {
				if schema.Ref == "" {
					continue
				}

				if len(op.Tags) == 0 {
					tag := "shared"
					schemasByTag[tag] = append(schemasByTag[tag], schema.Ref)
					schemaRefs[schema.Ref] = []string{}
				}

				for _, tag := range op.Tags {
					tagLower := strings.ToLower(tag)
					if !slices.Contains(schemasByTag[tagLower], schema.Ref) {
						schemasByTag[tagLower] = append(schemasByTag[tagLower], schema.Ref)
					}
					if !slices.Contains(schemaRefs[schema.Ref], tagLower) {
						schemaRefs[schema.Ref] = append(schemaRefs[schema.Ref], tagLower)
					}
				}
			}
		}
	}

	// Filter out the schemas that are referenced from multiple tags
	for schema, refs := range schemaRefs {
		if len(refs) > 1 {
			for _, tag := range refs {
				tagLower := strings.ToLower(tag)
				schemasByTag[tagLower] = remove(
					schemasByTag[tagLower],
					slices.Index(schemasByTag[tagLower], schema),
				)
			}
			schemasByTag["shared"] = append(schemasByTag["shared"], schema)
		}
	}

	for _, schemas := range schemasByTag {
		slices.Sort(schemas)
	}

	b.schemasByTag = schemasByTag
}

func (b *Builder) collectResponses() {
	responsesByTag := make(map[string][]string)
	tagsByResponse := make(map[string][]string)

	for _, pathItem := range b.spec.Paths.Map() {
		for _, op := range pathItem.Operations() {
			if op.Responses == nil {
				continue
			}

			schemas := make([]*openapi3.ResponseRef, 0, op.Responses.Len())
			for code, response := range op.Responses.Map() {
				if response.Ref != "" {
					schemas = append(schemas, response)
					if code == "default" || !strings.HasPrefix(code, "2") {
						b.errorSchemas[response.Ref] = struct{}{}
					}
				}
			}

			for _, schema := range schemas {
				if schema.Ref == "" {
					continue
				}

				if len(op.Tags) == 0 {
					tag := "shared"
					responsesByTag[tag] = append(responsesByTag[tag], schema.Ref)
					tagsByResponse[schema.Ref] = []string{}
				}

				for _, tag := range op.Tags {
					tagLower := strings.ToLower(tag)
					if !slices.Contains(responsesByTag[tagLower], schema.Ref) {
						responsesByTag[tagLower] = append(responsesByTag[tagLower], schema.Ref)
					}
					if !slices.Contains(tagsByResponse[schema.Ref], tagLower) {
						tagsByResponse[schema.Ref] = append(tagsByResponse[schema.Ref], tagLower)
					}
				}
			}
		}
	}

	// Filter out the schemas that are referenced from multiple tags
	for schema, refs := range tagsByResponse {
		if len(refs) > 1 {
			for _, tag := range refs {
				tagLower := strings.ToLower(tag)
				responsesByTag[tagLower] = remove(
					responsesByTag[tagLower],
					slices.Index(responsesByTag[tagLower], schema),
				)
			}
			responsesByTag["shared"] = append(responsesByTag["shared"], schema)
		}
	}

	for _, schemas := range responsesByTag {
		slices.Sort(schemas)
	}

	b.responsesByTag = responsesByTag
}

// Collect the schemas that are referenced in the response body of the given operation.
func (b *Builder) collectSchemasInResponse(op *openapi3.Operation) []*openapi3.SchemaRef {
	if op.Responses == nil || op.Responses.Len() == 0 {
		return nil
	}

	schemas := make([]*openapi3.SchemaRef, 0, op.Responses.Len())
	for code, response := range op.Responses.Map() {
		// TODO: handle content-type correctly
		for _, mediaType := range response.Value.Content {
			schema := mediaType.Schema
			if schema == nil {
				continue
			}
			if code == "default" || !strings.HasPrefix(code, "2") {
				b.errorSchemas[schema.Ref] = struct{}{}
			}
			schemas = append(schemas, schema)
		}
	}

	return collectReferencedSchemas(schemas)
}

func (b *Builder) tagByTagName(name string) *openapi3.Tag {
	idx := slices.IndexFunc(b.spec.Tags, func(tag *openapi3.Tag) bool {
		return strings.EqualFold(tag.Name, name)
	})
	tag := &openapi3.Tag{
		Name: name,
	}
	if idx != -1 {
		tag = b.spec.Tags[idx]
	}
	return tag
}

// Collect the schemas that are referenced in the request parameters of the given operation.
func collectSchemasInParams(op *openapi3.Operation) []*openapi3.SchemaRef {
	if len(op.Parameters) == 0 {
		return nil
	}

	schemas := make([]*openapi3.SchemaRef, 0, len(op.Parameters))

	// Iterate over the parameters of the operation
	for _, param := range op.Parameters {
		if param == nil || param.Value == nil {
			continue
		}
		schemas = append(schemas, param.Value.Schema)
	}

	return append(schemas, collectReferencedSchemas(schemas)...)
}

// Collect the schemas that are referenced in the request body of the given operation.
func collectSchemasInRequest(op *openapi3.Operation) []*openapi3.SchemaRef {
	if op.RequestBody == nil {
		return nil
	}

	schemas := make([]*openapi3.SchemaRef, 0, len(op.RequestBody.Value.Content))

	// Iterate over the responses of the operation
	for _, mediaType := range op.RequestBody.Value.Content {
		if mediaType.Schema == nil {
			continue
		}
		schemas = append(schemas, mediaType.Schema)
	}

	return append(schemas, collectReferencedSchemas(schemas)...)
}

// Collect the schemas that are referenced by the given schemas
func collectReferencedSchemas(schemas []*openapi3.SchemaRef) []*openapi3.SchemaRef {
	referencedSchemasMap := make(map[string]*openapi3.SchemaRef)

	// Iterate over the given schemas
	for _, schema := range schemas {
		// Collect the referenced schemas recursively
		collectReferencedSchemasRecursive(schema, referencedSchemasMap)
	}

	return slices.Collect(maps.Values(referencedSchemasMap))
}

// Recursively collect the schemas that are referenced by the given schema
func collectReferencedSchemasRecursive(
	schema *openapi3.SchemaRef,
	referencedSchemasMap map[string]*openapi3.SchemaRef,
) {
	if schema == nil {
		return
	}

	// save referenced schemas for later lookup
	if schema.Ref != "" {
		referencedSchemasMap[schema.Ref] = schema
	}

	if schema.Value.Type.Is("object") {
		for _, prop := range schema.Value.Properties {
			collectReferencedSchemasRecursive(prop, referencedSchemasMap)
		}
	}

	if schema.Value.Type.Is("array") && schema.Value.Items != nil {
		collectReferencedSchemasRecursive(schema.Value.Items, referencedSchemasMap)
	}

	if schema.Value.AnyOf != nil {
		for _, one := range schema.Value.AnyOf {
			collectReferencedSchemasRecursive(one, referencedSchemasMap)
		}
	}

	if schema.Value.AllOf != nil {
		for _, one := range schema.Value.AllOf {
			collectReferencedSchemasRecursive(one, referencedSchemasMap)
		}
	}

	if schema.Value.OneOf != nil {
		for _, one := range schema.Value.OneOf {
			collectReferencedSchemasRecursive(one, referencedSchemasMap)
		}
	}
}

func remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}
