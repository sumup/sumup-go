package main

import (
	"fmt"
	"os"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func loadOpenAPIDocument(filename string) (*v3.Document, error) {
	spec, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read specs: %w", err)
	}

	document, err := libopenapi.NewDocument(spec)
	if err != nil {
		return nil, fmt.Errorf("load openapi document: %w", err)
	}

	model, err := document.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("build openapi v3 model: %w", err)
	}

	return &model.Model, nil
}
