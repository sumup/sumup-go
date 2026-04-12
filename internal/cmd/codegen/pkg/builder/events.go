package builder

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"
	"github.com/sumup/sumup-go/internal/cmd/codegen/pkg/extension"
)

type eventDefinition struct {
	Name   string
	Type   string
	Object string
}

func (b *Builder) eventDefinitions() ([]eventDefinition, error) {
	var definitions []eventDefinition
	if b.spec.Webhooks == nil {
		return definitions, nil
	}
	names := make(map[string]string)
	for eventType, item := range b.spec.Webhooks.FromOldest() {
		if item == nil || item.Post == nil {
			return nil, fmt.Errorf("webhook %s: missing POST operation", eventType)
		}
		op := item.Post
		if op.OperationId == "" {
			return nil, fmt.Errorf("webhook %s: missing operation ID", eventType)
		}
		name := strcase.ToCamel(strings.TrimSuffix(op.OperationId, "Webhook"))
		if previous, exists := names[name]; exists {
			return nil, fmt.Errorf("webhooks %s and %s have the same name %s", previous, eventType, name)
		}
		object, ok := extension.Get[map[string]string](op.Extensions, "x-object")
		if !ok || !strings.HasPrefix(object["$ref"], "#/components/schemas/") {
			return nil, fmt.Errorf("webhook %s: x-object must reference a local schema", eventType)
		}
		ref := object["$ref"]
		if b.spec.Components == nil || b.spec.Components.Schemas == nil {
			return nil, fmt.Errorf("webhook %s: cannot resolve object %s", eventType, ref)
		}
		if _, ok := b.spec.Components.Schemas.Get(strings.TrimPrefix(ref, "#/components/schemas/")); !ok {
			return nil, fmt.Errorf("webhook %s: cannot resolve object %s", eventType, ref)
		}
		names[name] = eventType
		definitions = append(definitions, eventDefinition{Name: name, Type: eventType, Object: b.schemaTypeName(ref)})
	}
	slices.SortFunc(definitions, func(a, b eventDefinition) int { return strings.Compare(a.Type, b.Type) })
	return definitions, nil
}

func (b *Builder) writeEventsFile() error {
	definitions, err := b.eventDefinitions()
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := b.templates.ExecuteTemplate(&buf, "events.go.tmpl", definitions); err != nil {
		return fmt.Errorf("render events: %w", err)
	}
	source, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format events: %w", err)
	}
	if err := os.MkdirAll(b.cfg.Out, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(b.cfg.Out, "events_generated.go"), source, 0o644); err != nil {
		return fmt.Errorf("write events: %w", err)
	}
	return nil
}
