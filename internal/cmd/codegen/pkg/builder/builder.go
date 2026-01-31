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

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	v3low "github.com/pb33f/libopenapi/datamodel/low/v3"

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
	spec *v3.Document

	// schemasByTag maps tags to respective schema references.
	schemasByTag    map[string][]string
	resolvedSchemas map[string][]NamedSchema

	responsesByTag    map[string][]string
	resolvedResponses map[string][]*v3.Response

	// errorSchemas are refs of schemas that are used for error responses (status code >= 400).
	errorSchemas map[string]struct{}
	// schemaNameOverrides maps schema refs to overridden type names.
	schemaNameOverrides map[string]string
	pathsByTag          map[string]*v3.Paths

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
		resolvedSchemas:     make(map[string][]NamedSchema),
		resolvedResponses:   make(map[string][]*v3.Response),
		pathsByTag:          make(map[string]*v3.Paths),
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
func (b *Builder) Load(spec *v3.Document) error {
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
		b.pathsByTag["shared"] = v3.NewPaths(&v3low.Paths{})
	}

	for tagName, paths := range b.pathsByTag {
		if err := b.generateResource(tagName, paths); err != nil {
			return err
		}
	}

	if err := b.writeClientFile(path.Join(b.cfg.Out, "client.go"), slices.Collect(maps.Keys(b.pathsByTag))); err != nil {
		return err
	}

	if err := b.writeAPIVersionFile(); err != nil {
		return err
	}

	took := time.Since(b.start)
	slog.Info("sdk generated", slog.Duration("took", took))

	return nil
}

func (b *Builder) collectPaths() {
	for path, pathSpec := range b.spec.Paths.PathItems.FromOldest() {
		for method, operationSpec := range pathSpec.GetOperations().FromOldest() {
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
				tagPaths = v3.NewPaths(&v3low.Paths{})
				b.pathsByTag[tag] = tagPaths
			}

			tagPaths.PathItems.Set(path, pathSpec)
		}
	}
}

func (b *Builder) resolveResponses() {
	if b.spec == nil || b.spec.Components == nil || b.spec.Components.Responses == nil {
		return
	}

	for tag, refs := range b.responsesByTag {
		for _, ref := range refs {
			name := strings.TrimPrefix(ref, "#/components/responses/")
			resolved, ok := b.spec.Components.Responses.Get(name)
			if !ok {
				slog.Warn("response reference not found",
					slog.String("ref", ref),
					slog.String("tag", tag),
				)
				continue
			}
			resolved.Reference = ref
			b.resolvedResponses[tag] = append(b.resolvedResponses[tag], resolved)
		}
	}
}

func (b *Builder) resolveSchemas() error {
	for tag, refs := range b.schemasByTag {
		for _, ref := range refs {
			resolved, err := b.resolveSchemaRef(ref)
			if err != nil {
				return fmt.Errorf("resolve reference %q: %w", ref, err)
			}

			b.resolvedSchemas[tag] = append(b.resolvedSchemas[tag], NamedSchema{
				Ref:    ref,
				Schema: resolved,
			})
		}
	}

	return nil
}

func (b *Builder) collectSchemas() {
	// Map of schemas grouped by tag
	schemasByTag := make(map[string][]string)

	// Map of schemas to list of tags that reference them
	schemaRefs := make(map[string][]string)

	for _, pathItem := range b.spec.Paths.PathItems.FromOldest() {
		for _, op := range pathItem.GetOperations().FromOldest() {
			operationName := strcase.ToCamel(op.OperationId)
			methodName := operationMethodName(op)

			requestSchemas := collectSchemasInRequest(op)
			schemas := b.collectSchemasInResponse(op)
			schemas = append(schemas, collectSchemasInParams(op)...)
			schemas = append(schemas, requestSchemas...)
			if methodName != operationName {
				for _, schema := range requestSchemas {
					if schema == nil || schema.GetReference() == "" {
						continue
					}

					refName := strings.TrimPrefix(schema.GetReference(), "#/components/schemas/")
					if !strings.HasPrefix(refName, operationName) {
						continue
					}

					suffix := strings.TrimPrefix(refName, operationName)
					if suffix == "" {
						continue
					}

					b.schemaNameOverrides[schema.GetReference()] = methodName + suffix
				}
			}

			for _, schema := range schemas {
				if schema == nil || schema.GetReference() == "" {
					continue
				}

				ref := schema.GetReference()
				if len(op.Tags) == 0 {
					tag := "shared"
					schemasByTag[tag] = append(schemasByTag[tag], ref)
					schemaRefs[ref] = []string{}
				}

				for _, tag := range op.Tags {
					tagLower := strings.ToLower(tag)
					if !slices.Contains(schemasByTag[tagLower], ref) {
						schemasByTag[tagLower] = append(schemasByTag[tagLower], ref)
					}
					if !slices.Contains(schemaRefs[ref], tagLower) {
						schemaRefs[ref] = append(schemaRefs[ref], tagLower)
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

	for _, pathItem := range b.spec.Paths.PathItems.FromOldest() {
		for _, op := range pathItem.GetOperations().FromOldest() {
			if op.Responses == nil {
				continue
			}

			schemas := make([]*v3.Response, 0, op.Responses.Codes.Len())
			for code, response := range op.Responses.Codes.FromOldest() {
				if response.Reference != "" {
					schemas = append(schemas, response)
					if code == "default" || !strings.HasPrefix(code, "2") {
						b.errorSchemas[response.Reference] = struct{}{}
					}
				}
			}

			for _, schema := range schemas {
				if schema == nil || schema.Reference == "" {
					continue
				}

				if len(op.Tags) == 0 {
					tag := "shared"
					responsesByTag[tag] = append(responsesByTag[tag], schema.Reference)
					tagsByResponse[schema.Reference] = []string{}
				}

				for _, tag := range op.Tags {
					tagLower := strings.ToLower(tag)
					if !slices.Contains(responsesByTag[tagLower], schema.Reference) {
						responsesByTag[tagLower] = append(responsesByTag[tagLower], schema.Reference)
					}
					if !slices.Contains(tagsByResponse[schema.Reference], tagLower) {
						tagsByResponse[schema.Reference] = append(tagsByResponse[schema.Reference], tagLower)
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
func (b *Builder) collectSchemasInResponse(op *v3.Operation) []*base.SchemaProxy {
	if op.Responses == nil || op.Responses.Codes.Len() == 0 {
		return nil
	}

	schemas := make([]*base.SchemaProxy, 0, op.Responses.Codes.Len())
	for code, response := range op.Responses.Codes.FromOldest() {
		// TODO: handle content-type correctly
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

		for _, mediaType := range response.Content.FromOldest() {
			schema := mediaType.Schema
			if schema == nil {
				continue
			}
			if code == "default" || !strings.HasPrefix(code, "2") {
				if schema.GetReference() != "" {
					b.errorSchemas[schema.GetReference()] = struct{}{}
				}
			}
			schemas = append(schemas, schema)
		}
	}

	return collectReferencedSchemas(schemas)
}

func (b *Builder) tagByTagName(name string) *base.Tag {
	idx := slices.IndexFunc(b.spec.Tags, func(tag *base.Tag) bool {
		return strings.EqualFold(tag.Name, name)
	})
	tag := &base.Tag{
		Name: name,
	}
	if idx != -1 {
		tag = b.spec.Tags[idx]
	}
	return tag
}

// Collect the schemas that are referenced in the request parameters of the given operation.
func collectSchemasInParams(op *v3.Operation) []*base.SchemaProxy {
	if len(op.Parameters) == 0 {
		return nil
	}

	schemas := make([]*base.SchemaProxy, 0, len(op.Parameters))

	// Iterate over the parameters of the operation
	for _, param := range op.Parameters {
		if param == nil || param.Schema == nil {
			continue
		}
		schemas = append(schemas, param.Schema)
	}

	return append(schemas, collectReferencedSchemas(schemas)...)
}

// Collect the schemas that are referenced in the request body of the given operation.
func collectSchemasInRequest(op *v3.Operation) []*base.SchemaProxy {
	if op.RequestBody == nil || op.RequestBody.Content == nil {
		return nil
	}

	schemas := make([]*base.SchemaProxy, 0, op.RequestBody.Content.Len())

	// Iterate over the responses of the operation
	for _, mediaType := range op.RequestBody.Content.FromOldest() {
		if mediaType.Schema == nil {
			continue
		}
		schemas = append(schemas, mediaType.Schema)
	}

	return append(schemas, collectReferencedSchemas(schemas)...)
}

// Collect the schemas that are referenced by the given schemas
func collectReferencedSchemas(schemas []*base.SchemaProxy) []*base.SchemaProxy {
	referencedSchemasMap := make(map[string]*base.SchemaProxy)

	// Iterate over the given schemas
	for _, schema := range schemas {
		// Collect the referenced schemas recursively
		collectReferencedSchemasRecursive(schema, referencedSchemasMap)
	}

	return slices.Collect(maps.Values(referencedSchemasMap))
}

// Recursively collect the schemas that are referenced by the given schema
func collectReferencedSchemasRecursive(
	schema *base.SchemaProxy,
	referencedSchemasMap map[string]*base.SchemaProxy,
) {
	if schema == nil {
		return
	}

	// save referenced schemas for later lookup
	if schema.GetReference() != "" {
		referencedSchemasMap[schema.GetReference()] = schema
	}

	inner := schema.Schema()
	if inner == nil {
		return
	}

	if slices.Contains(inner.Type, "object") && inner.Properties != nil {
		for _, prop := range inner.Properties.FromOldest() {
			collectReferencedSchemasRecursive(prop, referencedSchemasMap)
		}
	}

	if inner.Items != nil && inner.Items.IsA() && inner.Items.A != nil {
		collectReferencedSchemasRecursive(inner.Items.A, referencedSchemasMap)
	}

	if inner.AnyOf != nil {
		for _, one := range inner.AnyOf {
			collectReferencedSchemasRecursive(one, referencedSchemasMap)
		}
	}

	if inner.AllOf != nil {
		for _, one := range inner.AllOf {
			collectReferencedSchemasRecursive(one, referencedSchemasMap)
		}
	}

	if inner.OneOf != nil {
		for _, one := range inner.OneOf {
			collectReferencedSchemasRecursive(one, referencedSchemasMap)
		}
	}
}

func (b *Builder) resolveSchemaRef(ref string) (*base.SchemaProxy, error) {
	components := b.spec.Components
	if components == nil || components.Schemas == nil {
		return nil, fmt.Errorf("missing components.schemas")
	}

	name := strings.TrimPrefix(ref, "#/components/schemas/")
	resolved, ok := components.Schemas.Get(name)
	if !ok {
		return nil, fmt.Errorf("cannot resolve $ref: %s", ref)
	}

	return resolved, nil
}

func (b *Builder) resolveResponseRef(ref string) *v3.Response {
	components := b.spec.Components
	if components == nil || components.Responses == nil {
		return nil
	}

	name := strings.TrimPrefix(ref, "#/components/responses/")
	resolved, ok := components.Responses.Get(name)
	if !ok {
		return nil
	}

	return resolved
}

func (b *Builder) resolveParameter(param *v3.Parameter) *v3.Parameter {
	if param == nil {
		return nil
	}
	if !param.IsReference() {
		return param
	}

	components := b.spec.Components
	if components == nil || components.Parameters == nil {
		return nil
	}

	name := strings.TrimPrefix(param.GetReference(), "#/components/parameters/")
	resolved, ok := components.Parameters.Get(name)
	if !ok {
		return nil
	}

	return resolved
}

func remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}
