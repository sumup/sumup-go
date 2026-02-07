package ptr_test

import (
	"testing"

	"github.com/sumup/sumup-go/internal/ptr"
)

func TestOrNil(t *testing.T) {
	t.Parallel()

	t.Run("nil pointer returns nil", func(t *testing.T) {
		t.Parallel()

		got := ptr.OrNil[string](nil)
		if got != nil {
			t.Errorf("OrNil() = %v, want %v", got, nil)
		}
	})

	t.Run("non nil pointer returns value", func(t *testing.T) {
		t.Parallel()

		val := "hello"
		got := ptr.OrNil(&val)
		if got != val {
			t.Errorf("OrNil() = %v, want %v", got, val)
		}
	})
}
