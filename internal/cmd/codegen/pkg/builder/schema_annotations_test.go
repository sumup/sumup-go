package builder

import (
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
	data, err := os.ReadFile(filepath.Join(out, "merchants.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, snippet := range []string{"type CountryCode string", "Country CountryCode", "PhoneNumber PhoneNumber", "type ExtendedAddress struct", "Label *string"} {
		if !strings.Contains(string(data), snippet) {
			t.Errorf("missing %q", snippet)
		}
	}
}
