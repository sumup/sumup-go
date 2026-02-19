package builder

import (
	"testing"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

func TestGetJSONMediaType(t *testing.T) {
	t.Run("prefers application/json", func(t *testing.T) {
		content := orderedmap.New[string, *v3.MediaType]()
		jsonMT := &v3.MediaType{}
		problemMT := &v3.MediaType{}
		content.Set("application/problem+json", problemMT)
		content.Set("application/json", jsonMT)

		got, ok := getJSONMediaType(content)
		if !ok {
			t.Fatal("expected content type to be found")
		}
		if got != jsonMT {
			t.Fatalf("got %p, want %p", got, jsonMT)
		}
	})

	t.Run("supports application/problem+json", func(t *testing.T) {
		content := orderedmap.New[string, *v3.MediaType]()
		problemMT := &v3.MediaType{}
		content.Set("application/problem+json", problemMT)

		got, ok := getJSONMediaType(content)
		if !ok {
			t.Fatal("expected content type to be found")
		}
		if got != problemMT {
			t.Fatalf("got %p, want %p", got, problemMT)
		}
	})

	t.Run("returns false when no matching content type exists", func(t *testing.T) {
		content := orderedmap.New[string, *v3.MediaType]()
		content.Set("text/plain", &v3.MediaType{})

		got, ok := getJSONMediaType(content)
		if ok {
			t.Fatalf("got ok=%v, want ok=false", ok)
		}
		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})
}
