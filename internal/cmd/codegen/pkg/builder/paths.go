package builder

import (
	"slices"
	"sort"
	"strings"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// pathsInMatchingOrder mimics kin-openapi Paths.InMatchingOrder ordering.
func pathsInMatchingOrder(paths *v3.Paths) []string {
	if paths == nil || paths.PathItems == nil || paths.PathItems.Len() == 0 {
		return nil
	}

	vars := make(map[int][]string)
	max := 0
	for path := range paths.PathItems.KeysFromOldest() {
		count := strings.Count(path, "}")
		vars[count] = append(vars[count], path)
		if count > max {
			max = count
		}
	}

	ordered := make([]string, 0, paths.PathItems.Len())
	for c := 0; c <= max; c++ {
		if ps, ok := vars[c]; ok {
			slices.Sort(ps)
			sort.Sort(sort.Reverse(sort.StringSlice(ps)))
			ordered = append(ordered, ps...)
		}
	}

	return ordered
}
