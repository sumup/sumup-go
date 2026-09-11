package builder

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

func TestBuilder_writeEventsFile(t *testing.T) {
	t.Parallel()
	spec, err := os.ReadFile("../../../../../openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := libopenapi.NewDocument(spec)
	if err != nil {
		t.Fatal(err)
	}
	model, err := doc.BuildV3Model()
	if err != nil {
		t.Fatal(err)
	}
	b := New(Config{Out: t.TempDir()})
	if err := b.Load(&model.Model); err != nil {
		t.Fatal(err)
	}
	if err := b.writeEventsFile(); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(b.cfg.Out, "events_generated.go")
	first, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("typed callbacks from OpenAPI", func(t *testing.T) {
		t.Parallel()
		for _, expected := range []string{
			`const EventTypeMemberUpdated = "members.updated"`,
			`type MemberUpdatedEvent struct{ TypedEvent[Member] }`,
			`OnMemberUpdated(callback func(context.Context, *MemberUpdatedEvent) error) error`,
			`OnReaderCreated(callback func(context.Context, *ReaderCreatedEvent) error) error`,
			`return &MemberUpdatedEvent{TypedEvent[Member](raw)}`,
		} {
			if !strings.Contains(string(first), expected) {
				t.Errorf("missing %s", expected)
			}
		}
	})
	t.Run("deterministic output", func(t *testing.T) {
		t.Parallel()
		if err := b.writeEventsFile(); err != nil {
			t.Fatal(err)
		}
		second, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first, second) {
			t.Fatal("event generation is not deterministic")
		}
	})
}

func TestBuilder_eventDefinitions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, operation, ref, want string
		duplicate                  bool
	}{
		{"missing object", "ThingCreatedWebhook", "", "x-object", false},
		{"external object", "ThingCreatedWebhook", "other.json#/Thing", "x-object", false},
		{"missing schema", "ThingCreatedWebhook", "#/components/schemas/Missing", "resolve object", false},
		{"missing operation", "", "#/components/schemas/Thing", "operation ID", false},
		{"duplicate operation", "ThingCreatedWebhook", "#/components/schemas/Thing", "same name", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			op := &v3.Operation{OperationId: tc.operation}
			if tc.ref != "" {
				var object yaml.Node
				if err := object.Encode(map[string]string{"$ref": tc.ref}); err != nil {
					t.Fatal(err)
				}
				op.Extensions = orderedmap.New[string, *yaml.Node]()
				op.Extensions.Set("x-object", &object)
			}
			hooks := orderedmap.New[string, *v3.PathItem]()
			hooks.Set("a", &v3.PathItem{Post: op})
			if tc.duplicate {
				hooks.Set("b", &v3.PathItem{Post: op})
			}
			schemas := orderedmap.New[string, *base.SchemaProxy]()
			schemas.Set("Thing", base.CreateSchemaProxy(&base.Schema{Type: []string{"object"}}))
			b := &Builder{
				spec:         &v3.Document{Webhooks: hooks, Components: &v3.Components{Schemas: schemas}},
				schemasByTag: map[string][]string{"things": {"#/components/schemas/Thing"}},
			}
			if _, err := b.eventDefinitions(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %s", err, tc.want)
			}
		})
	}
}
