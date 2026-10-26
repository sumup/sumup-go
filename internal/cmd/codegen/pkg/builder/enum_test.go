package builder

import (
	"testing"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/libopenapi/datamodel/high/base"
)

func TestNullableEnum(t *testing.T) {
	t.Parallel()

	var values yaml.Node
	if err := yaml.Unmarshal([]byte(`["credit", "debit", null]`), &values); err != nil {
		t.Fatal(err)
	}
	declaration := createEnum(&base.Schema{
		Type: []string{"string", "null"},
		Enum: values.Content[0].Content,
	}, "CardType").(*EnumDeclaration[string])
	if len(declaration.Values) != 2 {
		t.Fatalf("enum values = %v, want only credit and debit", declaration.Values)
	}
	if declaration.Values[0].Name != "CardTypeCredit" || declaration.Values[1].Name != "CardTypeDebit" {
		t.Fatalf("unexpected enum constants: %v", declaration.Values)
	}
}
