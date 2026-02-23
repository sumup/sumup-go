package nullable

import (
	"encoding/json"
	"errors"
	"testing"
)

type Foo struct {
	X *Field[string] `json:"x,omitempty"`
	Y *Field[string] `json:"y,omitempty"`
	Z *Field[string] `json:"z,omitempty"`
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

func TestField_Null(t *testing.T) {
	var nilField *Field[string]
	if nilField.Null() {
		t.Fatalf("(*Field[string])(nil).Null() = true, want false")
	}

	if !Null[string]().Null() {
		t.Fatalf("Null[string]().Null() = false, want true")
	}

	if Value("foo").Null() {
		t.Fatalf("Value(\"foo\").Null() = true, want false")
	}
}

func TestField_Value(t *testing.T) {
	var nilField *Field[string]
	if got := nilField.Value(); got != nil {
		t.Fatalf("(*Field[string])(nil).Value() = %v, want nil", got)
	}

	if got := Null[string]().Value(); got != nil {
		t.Fatalf("Null[string]().Value() = %v, want nil", got)
	}

	got := Value("foo").Value()
	if got == nil || *got != "foo" {
		t.Fatalf("Value(\"foo\").Value() = %v, want pointer to \"foo\"", got)
	}
}

func TestField_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		field   Field[string]
		want    string
		wantErr error
	}{
		{
			name:  "value",
			field: *Value("foo"),
			want:  `"foo"`,
		},
		{
			name:  "null",
			field: *Null[string](),
			want:  `null`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.field.MarshalJSON()
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("MarshalJSON() error = %v, want %v", err, tc.wantErr)
			}

			if gotS := string(got); gotS != tc.want {
				t.Fatalf("MarshalJSON() = %q, want %q", gotS, tc.want)
			}
		})
	}
}

func TestField_UnmarshalJSON(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		var got Field[string]
		if err := got.UnmarshalJSON([]byte(`"foo"`)); err != nil {
			t.Fatalf("UnmarshalJSON() error = %v", err)
		}

		if got.Null() {
			t.Fatalf("UnmarshalJSON(\"foo\").Null() = true, want false")
		}

		val := got.Value()
		if val == nil || *val != "foo" {
			t.Fatalf("UnmarshalJSON(\"foo\").Value() = %v, want pointer to \"foo\"", val)
		}
	})

	t.Run("null resets value", func(t *testing.T) {
		got := *Value("foo")
		if err := got.UnmarshalJSON([]byte(`null`)); err != nil {
			t.Fatalf("UnmarshalJSON() error = %v", err)
		}

		if !got.Null() {
			t.Fatalf("UnmarshalJSON(null).Null() = false, want true")
		}

		if val := got.Value(); val != nil {
			t.Fatalf("UnmarshalJSON(null).Value() = %v, want nil", val)
		}

		if s := got.String(); s != "" {
			t.Fatalf("UnmarshalJSON(null).String() = %q, want empty string", s)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		var got Field[string]
		if err := got.UnmarshalJSON([]byte(`{`)); err == nil {
			t.Fatalf("UnmarshalJSON(invalid JSON) error = nil, want non-nil")
		}
	})
}

func TestField_String(t *testing.T) {
	if got, want := Value("foo").String(), "foo"; got != want {
		t.Fatalf("Value(\"foo\").String() = %q, want %q", got, want)
	}

	if got, want := Null[string]().String(), ""; got != want {
		t.Fatalf("Null[string]().String() = %q, want %q", got, want)
	}
}
