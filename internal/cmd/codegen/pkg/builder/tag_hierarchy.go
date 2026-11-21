package builder

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type resourceParent struct {
	tag   string
	label string
}

func readResourceParents(doc *v3.Document) (map[string]resourceParent, error) {
	if doc.Extensions == nil {
		return nil, nil
	}
	node, ok := doc.Extensions.Get("x-tagGroups")
	if !ok {
		return nil, nil
	}
	var groups []struct {
		Name string   `yaml:"name"`
		Tags []string `yaml:"tags"`
	}
	if err := node.Decode(&groups); err != nil {
		return nil, fmt.Errorf("decode x-tagGroups: %w", err)
	}
	labels := make(map[string]string, len(doc.Tags))
	for _, tag := range doc.Tags {
		var displayName string
		if tag.Extensions != nil {
			if label, ok := tag.Extensions.Get("x-displayName"); ok {
				if err := label.Decode(&displayName); err != nil {
					return nil, fmt.Errorf("decode x-displayName for %s: %w", tag.Name, err)
				}
			}
		}
		labels[tag.Name] = cmp.Or(displayName, tag.Name)
	}
	parents := map[string]resourceParent{}
	for _, group := range groups {
		if labels[group.Name] == "" {
			continue
		}
		parent := strings.ToLower(group.Name)
		for _, child := range group.Tags {
			key := strings.ToLower(child)
			if child == group.Name || labels[child] == "" || parents[key].tag != "" {
				continue
			}
			for ancestor := parent; ancestor != ""; ancestor = parents[ancestor].tag {
				if ancestor == key {
					return nil, fmt.Errorf("tag hierarchy contains a cycle at %s", child)
				}
			}
			parents[key] = resourceParent{tag: parent, label: labels[child]}
		}
	}
	return parents, nil
}

func pathForResource(path *v3.PathItem, tag string) *v3.PathItem {
	operation := func(op *v3.Operation) *v3.Operation {
		if op == nil {
			return nil
		}
		owner := "shared"
		if len(op.Tags) > 0 {
			owner = strings.ToLower(op.Tags[0])
		}
		if owner != tag {
			return nil
		}
		copy := *op
		copy.Parameters = slices.Clone(op.Parameters)
		return &copy
	}
	return &v3.PathItem{
		Parameters: path.Parameters,
		Get:        operation(path.Get), Put: operation(path.Put), Post: operation(path.Post),
		Delete: operation(path.Delete), Options: operation(path.Options), Head: operation(path.Head),
		Patch: operation(path.Patch), Trace: operation(path.Trace),
	}
}

type childResource struct {
	Name    string
	Service string
}

func (b *Builder) children(tag string) []childResource {
	var children []childResource
	for child, parent := range b.resourceParents {
		if parent.tag == tag && b.pathsByTag[child] != nil {
			name := b.tagByTagName(child).Name
			children = append(children, childResource{Name: strcase.ToCamel(parent.label), Service: strcase.ToCamel(name)})
		}
	}
	slices.SortFunc(children, func(a, b childResource) int { return cmp.Compare(a.Name, b.Name) })
	return children
}

func (b *Builder) resourcePath(tag string) string {
	if parent := b.resourceParents[tag]; parent.tag != "" {
		return b.resourcePath(parent.tag) + "." + strcase.ToCamel(parent.label)
	}
	return strcase.ToCamel(tag)
}
