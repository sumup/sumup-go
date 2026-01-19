package builder

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"
)

type Alias struct {
	Name    string
	Comment string
}

type templateData struct {
	PackageName string
	Aliases     []Alias
	Types       []Writable
	Service     string
	Methods     []*Method
}

func (b *Builder) generateResource(tagName string, paths *openapi3.Paths) error {
	if tagName == "" {
		return fmt.Errorf("empty tag name")
	}

	resolvedSchemas := b.resolvedSchemas[tagName]
	resolvedResponses := b.resolvedResponses[tagName]

	tag := b.tagByTagName(tagName)

	types := b.schemasToTypes(resolvedSchemas, b.errorSchemas)

	bodyTypes := b.pathsToBodyTypes(tagName, paths)
	types = append(types, bodyTypes...)

	paramTypes := b.pathsToParamTypes(tagName, paths)
	types = append(types, paramTypes...)

	responseTypes := b.pathsToResponseTypes(tagName, paths)
	types = append(types, responseTypes...)

	respTypes := b.respToTypes(resolvedResponses, b.errorSchemas)
	types = append(types, respTypes...)

	methods, err := b.pathsToMethods(tagName, paths)
	if err != nil {
		return fmt.Errorf("convert paths to methods: %w", err)
	}

	slog.Info("generating file",
		slog.String("tag", tag.Name),
		slog.Int("schema_structs", len(types)),
		slog.Int("body_structs", len(bodyTypes)),
		slog.Int("path_params_structs", len(paramTypes)),
		slog.Int("response_structs", len(respTypes)),
	)

	if err := os.MkdirAll(b.cfg.Out, os.ModePerm); err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	if err := b.templates.ExecuteTemplate(buf, "resource.go.tmpl", templateData{
		PackageName: strcase.ToSnake(tag.Name),
		Types:       types,
		Methods:     methods,
		Service:     strcase.ToCamel(tag.Name),
	}); err != nil {
		return err
	}

	fName := path.Join(b.cfg.Out, fmt.Sprintf("%s.go", strcase.ToSnake(tag.Name)))
	f, err := openGeneratedFile(fName)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(buf.String()); err != nil {
		return err
	}

	return nil
}

func (b *Builder) writeClientFile(fname string, tags []string) error {
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(0o755))
	if err != nil {
		return fmt.Errorf("create %q: %w", fname, err)
	}
	defer func() { _ = f.Close() }()

	type resource struct {
		Name       string
		ClientName string
	}

	resources := make([]resource, 0, len(tags))
	for i := range tags {
		if p := b.pathsByTag[tags[i]]; p.Len() == 0 {
			continue
		}
		resources = append(resources, resource{
			Name:       strcase.ToCamel(tags[i]),
			ClientName: strcase.ToCamel(tags[i]) + "Client",
		})
	}

	slices.SortFunc(resources, func(a, b resource) int {
		return strings.Compare(a.Name, b.Name)
	})

	if err := b.templates.ExecuteTemplate(f, "base.go.tmpl", map[string]any{
		"Resources": resources,
	}); err != nil {
		return fmt.Errorf("generate client: %w", err)
	}

	return nil
}

func (b *Builder) writeAPIVersionFile() error {
	if b.spec == nil || b.spec.Info == nil {
		return fmt.Errorf("missing specs info: call Load to load the specs first")
	}

	apiVersion := strings.TrimSpace(b.spec.Info.Version)
	if apiVersion == "" {
		return fmt.Errorf("missing api version in spec info")
	}

	dir := path.Join(b.cfg.Out, "internal")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("create %q: %w", dir, err)
	}

	fName := path.Join(dir, "apiversion.go")
	f, err := openGeneratedFile(fName)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := fmt.Fprintf(f, "package internal\n\nconst APIVersion = %q\n", apiVersion); err != nil {
		return fmt.Errorf("write %q: %w", fName, err)
	}

	return nil
}

func openGeneratedFile(filename string) (*os.File, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get current working directory: %w", err)
	}

	p := filepath.Join(cwd, filename)
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(0o755))
	if err != nil {
		return nil, fmt.Errorf("create %q: %w", p, err)
	}

	return f, nil
}
