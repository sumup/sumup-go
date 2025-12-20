package nullable

import (
	"encoding/json"
	"testing"
)

type Foo struct {
	X Field[string] `json:"x,omitzero"`
	Y Field[string] `json:"y,omitzero"`
	Z Field[string] `json:"z,omitzero"`
}

func TestField(t *testing.T) {
	foo := Foo{
		X: Value("foo"),
		Y: Null[string](),
	}

	data, err := json.Marshal(&foo)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if got, want := string(data), `{"x":"foo","y":null}`; got != want {
		t.Fatalf("Marshal() = %q, want %q", got, want)
	}
}
