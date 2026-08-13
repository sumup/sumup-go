package builder

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
)

func TestBuilderSamples(t *testing.T) {
	t.Parallel()

	repositoryRoot, catalog, expectedSamples := testSampleCatalog(t)
	if catalog.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", catalog.SchemaVersion)
	}
	if catalog.SDK.Module != "github.com/sumup/sumup-go" {
		t.Fatalf("SDK.Module = %q", catalog.SDK.Module)
	}
	if catalog.Language != "go" {
		t.Fatalf("Language = %q, want go", catalog.Language)
	}
	if catalog.OpenAPIVersion != "1.0.0" {
		t.Fatalf("OpenAPIVersion = %q, want 1.0.0", catalog.OpenAPIVersion)
	}
	if len(catalog.Samples) != expectedSamples {
		t.Fatalf("len(Samples) = %d, want %d", len(catalog.Samples), expectedSamples)
	}
	if !slices.IsSortedFunc(catalog.Samples, func(a, b Sample) int {
		return strings.Compare(a.ID, b.ID)
	}) {
		t.Fatal("samples are not sorted by ID")
	}
	seen := make(map[string]struct{}, len(catalog.Samples))
	for _, sample := range catalog.Samples {
		if _, ok := seen[sample.ID]; ok {
			t.Fatalf("duplicate sample ID %q", sample.ID)
		}
		seen[sample.ID] = struct{}{}
	}

	createCheckout := sampleByID(t, catalog.Samples, "CreateCheckout.HostedCheckout")
	if !strings.Contains(createCheckout.Source, "sumup.CheckoutsCreateParams{") {
		t.Fatalf("CreateCheckout sample does not use the generated params type:\n%s", createCheckout.Source)
	}
	if !strings.Contains(createCheckout.Source, `CheckoutReference: "b50pr914-6k0e-3091-a592-890010285b3d"`) {
		t.Fatalf("CreateCheckout sample does not use the OpenAPI example:\n%s", createCheckout.Source)
	}
	if !strings.Contains(createCheckout.Source, `new("A sample checkout")`) {
		t.Fatalf("CreateCheckout sample does not use new for a pointer value:\n%s", createCheckout.Source)
	}
	if !strings.Contains(createCheckout.Source, `HostedCheckout: new(sumup.HostedCheckout{`) {
		t.Fatalf("CreateCheckout sample does not use new for a pointer to a struct:\n%s", createCheckout.Source)
	}
	checkout := sampleByID(t, catalog.Samples, "CreateCheckout.Checkout")
	if !strings.Contains(checkout.Source, `time.Date(2020, time.February, 29, 10, 56, 56, 0, time.UTC)`) {
		t.Fatalf("CreateCheckout sample does not use time.Date for timestamps:\n%s", checkout.Source)
	}
	if strings.Contains(checkout.Source, "time.Parse") || strings.Contains(checkout.Source, "func mustParseTime") {
		t.Fatalf("CreateCheckout sample parses timestamps at runtime:\n%s", checkout.Source)
	}
	encodedSample, err := json.Marshal(createCheckout)
	if err != nil {
		t.Fatalf("marshal CreateCheckout sample: %v", err)
	}
	if !strings.Contains(string(encodedSample), `"sample":`) {
		t.Fatalf("sample JSON does not preserve the portal contract: %s", encodedSample)
	}
	if strings.Contains(string(encodedSample), `"source":`) {
		t.Fatalf("sample JSON contains internal source field name: %s", encodedSample)
	}

	for _, sample := range catalog.Samples {
		if strings.Contains(sample.Source, "ptr(") || strings.Contains(sample.Source, "func ptr[") {
			t.Errorf("sample %q contains the obsolete ptr helper:\n%s", sample.ID, sample.Source)
		}
		if _, err := parser.ParseFile(token.NewFileSet(), sample.ID+".go", sample.Source, parser.AllErrors); err != nil {
			t.Errorf("parse sample %q: %v", sample.ID, err)
		}
	}

	compileSamples(t, repositoryRoot, catalog.Samples)
}

func TestBuilderSamplesDeterministic(t *testing.T) {
	t.Parallel()

	_, first, _ := testSampleCatalog(t)
	_, second, _ := testSampleCatalog(t)
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first catalog: %v", err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("marshal second catalog: %v", err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatal("sample generation is not deterministic")
	}
}

func testSampleCatalog(t *testing.T) (string, *SampleCatalog, int) {
	t.Helper()

	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", "..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	spec, err := os.ReadFile(filepath.Join(repositoryRoot, "openapi.json"))
	if err != nil {
		t.Fatalf("read OpenAPI document: %v", err)
	}
	document, err := libopenapi.NewDocument(spec)
	if err != nil {
		t.Fatalf("load OpenAPI document: %v", err)
	}
	model, err := document.BuildV3Model()
	if err != nil {
		t.Fatalf("build OpenAPI model: %v", err)
	}

	builder := New(Config{})
	if err := builder.Load(&model.Model); err != nil {
		t.Fatalf("load builder: %v", err)
	}
	catalog, err := builder.Samples("test")
	if err != nil {
		t.Fatalf("generate samples: %v", err)
	}
	expectedSamples := 0
	for _, pathItem := range model.Model.Paths.PathItems.FromOldest() {
		for _, operation := range pathItem.GetOperations().FromOldest() {
			expectedSamples += len(requestExamples(operation))
		}
	}
	return repositoryRoot, catalog, expectedSamples
}

func sampleByID(t *testing.T, samples []Sample, id string) Sample {
	t.Helper()
	for _, sample := range samples {
		if sample.ID == id {
			return sample
		}
	}
	t.Fatalf("sample %q not found", id)
	return Sample{}
}

func compileSamples(t *testing.T, repositoryRoot string, samples []Sample) {
	t.Helper()

	dir := t.TempDir()
	module := fmt.Sprintf(`module generated-samples

go 1.26.0

require github.com/sumup/sumup-go v0.0.0

replace github.com/sumup/sumup-go => %s
`, repositoryRoot)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(module), 0o600); err != nil {
		t.Fatalf("write sample module: %v", err)
	}
	for i, sample := range samples {
		packageDir := filepath.Join(dir, fmt.Sprintf("sample%03d", i))
		if err := os.Mkdir(packageDir, 0o700); err != nil {
			t.Fatalf("create package for %q: %v", sample.ID, err)
		}
		if err := os.WriteFile(filepath.Join(packageDir, "main.go"), []byte(sample.Source), 0o600); err != nil {
			t.Fatalf("write sample %q: %v", sample.ID, err)
		}
	}

	command := exec.CommandContext(t.Context(), "go", "test", "-mod=mod", "./...")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("compile generated samples: %v\n%s", err, output)
	}
}
