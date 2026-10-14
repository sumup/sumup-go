package builder

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
)

func TestGenerateAnnotatedSchemas(t *testing.T) {
	t.Parallel()
	spec, err := os.ReadFile("testdata/schema-annotations.json")
	if err != nil {
		t.Fatal(err)
	}
	document, err := libopenapi.NewDocument(spec)
	if err != nil {
		t.Fatal(err)
	}
	model, err := document.BuildV3Model()
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relativeOut, err := filepath.Rel(cwd, out)
	if err != nil {
		t.Fatal(err)
	}
	generator := New(Config{Out: relativeOut})
	if err := generator.Load(&model.Model); err != nil {
		t.Fatal(err)
	}
	if err := generator.Build(); err != nil {
		t.Fatal(err)
	}
	generated, err := parser.ParseFile(token.NewFileSet(), filepath.Join(out, "merchants.go"), nil, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	var fields []string
	ast.Inspect(generated, func(node ast.Node) bool {
		declaration, ok := node.(*ast.TypeSpec)
		if !ok || declaration.Name.Name != "ExtendedAddress" {
			return true
		}
		object, ok := declaration.Type.(*ast.StructType)
		if !ok {
			t.Fatal("ExtendedAddress must be a struct")
		}
		for _, field := range object.Fields.List {
			for _, name := range field.Names {
				fields = append(fields, name.Name)
			}
		}
		return false
	})
	if strings.Join(fields, ",") != "City,Country,Label" {
		t.Fatalf("inherited fields = %v", fields)
	}
	checks := map[string][]string{
		"merchants.go": {"type CountryCode string", "Country CountryCode", "PhoneNumber PhoneNumber", "type ExtendedAddress struct", "Label *string"},
	}
	for file, snippets := range checks {
		data, err := os.ReadFile(filepath.Join(out, file))
		if err != nil {
			t.Fatal(err)
		}
		for _, snippet := range snippets {
			if !strings.Contains(string(data), snippet) {
				t.Errorf("%s missing %q:\n%s", file, snippet, data)
			}
		}
	}
}

func TestAnnotationNormalizationPreservesConstraints(t *testing.T) {
	t.Parallel()
	document, err := libopenapi.NewDocument([]byte(`{"openapi":"3.1.0","info":{"title":"test","version":"1"},"paths":{},"components":{"schemas":{"Code":{"type":"string"},"Constrained":{"allOf":[{"$ref":"#/components/schemas/Code"},{"minLength":2}]}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	model, err := document.BuildV3Model()
	if err != nil {
		t.Fatal(err)
	}
	normalizeSchemaAnnotations(&model.Model)
	constrained := model.Model.Components.Schemas.GetOrZero("Constrained").Schema()
	if len(constrained.AllOf) != 2 {
		t.Fatal("normalization removed a validation constraint")
	}
}

func TestAnnotatedRecursiveReferences(t *testing.T) {
	t.Parallel()
	document, err := libopenapi.NewDocument([]byte(`{"openapi":"3.1.0","info":{"title":"test","version":"1"},"paths":{"/nodes":{"get":{"tags":["Nodes"],"operationId":"GetNode","responses":{"200":{"description":"OK","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Node"}}}}}}}},"components":{"schemas":{"Node":{"type":"object","properties":{"next":{"$ref":"#/components/schemas/Node","description":"Next node."}}}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	model, err := document.BuildV3Model()
	if err != nil {
		t.Fatal(err)
	}
	generator := New(Config{})
	if err := generator.Load(&model.Model); err != nil {
		t.Fatal(err)
	}
}

func TestAnnotationNormalizationPreservesNamedReferences(t *testing.T) {
	t.Parallel()
	document, err := libopenapi.NewDocument([]byte(`{"openapi":"3.1.0","info":{"title":"test","version":"1"},"paths":{},"components":{"schemas":{"Code":{"type":"string"},"Alias":{"allOf":[{"description":"Alias documentation."},{"$ref":"#/components/schemas/Code"}]},"Holder":{"type":"object","properties":{"code":{"$ref":"#/components/schemas/Alias"}}}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	model, err := document.BuildV3Model()
	if err != nil {
		t.Fatal(err)
	}
	normalizeSchemaAnnotations(&model.Model)
	property := model.Model.Components.Schemas.GetOrZero("Holder").Schema().Properties.GetOrZero("code")
	if got := property.GetReference(); got != "#/components/schemas/Alias" {
		t.Fatalf("reference = %q, want Alias", got)
	}
	if got := property.Schema().Description; got != "Alias documentation." {
		t.Fatalf("description = %q", got)
	}
}

func TestAnnotationNormalizationPreservesObjectDependencies(t *testing.T) {
	t.Parallel()
	document, err := libopenapi.NewDocument([]byte(`{"openapi":"3.1.0","info":{"title":"test","version":"1"},"paths":{},"components":{"schemas":{"Base":{"type":"object","properties":{"id":{"type":"string"}}},"Derived":{"allOf":[{"$ref":"#/components/schemas/Base","description":"Inherited fields."}]}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	model, err := document.BuildV3Model()
	if err != nil {
		t.Fatal(err)
	}
	normalizeSchemaAnnotations(&model.Model)
	derived := model.Model.Components.Schemas.GetOrZero("Derived").Schema()
	if len(derived.AllOf) != 1 || derived.AllOf[0].GetReference() != "#/components/schemas/Base" {
		t.Fatal("object dependency was removed")
	}
}
