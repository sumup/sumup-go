package builder

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

func getJSONMediaType(content *orderedmap.Map[string, *v3.MediaType]) (*v3.MediaType, bool) {
	if content == nil {
		return nil, false
	}

	for _, mediaType := range []string{
		"application/json",
		"application/problem+json",
	} {
		mt, ok := content.Get(mediaType)
		if ok && mt != nil {
			return mt, true
		}
	}

	return nil, false
}
