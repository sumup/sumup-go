package main

import (
	"fmt"
	"os"

	"github.com/pb33f/libopenapi"

	"github.com/sumup/sumup-go/internal/cmd/codegen/pkg/builder"
)

func loadBuilder(specs, out string) (*builder.Builder, error) {
	spec, err := os.ReadFile(specs)
	if err != nil {
		return nil, fmt.Errorf("read specs: %w", err)
	}

	doc, err := libopenapi.NewDocument(spec)
	if err != nil {
		return nil, fmt.Errorf("load openapi document: %w", err)
	}

	model, err := doc.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("build openapi v3 model: %w", err)
	}

	b := builder.New(builder.Config{
		Out: out,
	})

	if err := b.Load(&model.Model); err != nil {
		return nil, fmt.Errorf("load spec: %w", err)
	}

	return b, nil
}
