package nullable

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	assert.Equal(t, `{"x":"foo","y":null}`, string(data))
}
