package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/pb33f/libopenapi"
	"github.com/urfave/cli/v2"

	"github.com/sumup/sumup-go/internal/cmd/codegen/pkg/builder"
)

func Generate() *cli.Command {
	var out string
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate SDK",
		Args:  true,
		Action: func(c *cli.Context) error {
			if !c.Args().Present() {
				return fmt.Errorf("empty argument, path to openapi specs expected")
			}

			specs := c.Args().First()

			if err := os.MkdirAll(out, os.ModePerm); err != nil {
				return fmt.Errorf("create output directory %q: %w", out, err)
			}

			spec, err := os.ReadFile(specs)
			if err != nil {
				return fmt.Errorf("read specs: %w", err)
			}

			doc, err := libopenapi.NewDocument(spec)
			if err != nil {
				return fmt.Errorf("load openapi document: %w", err)
			}

			model, err := doc.BuildV3Model()
			if err != nil {
				return fmt.Errorf("build openapi v3 model: %w", err)
			}

			builder := builder.New(builder.Config{
				Out: out,
			})

			if err := builder.Load(&model.Model); err != nil {
				return fmt.Errorf("load spec: %w", err)
			}

			if err := builder.Build(); err != nil {
				return fmt.Errorf("build sdk: %w", err)
			}

			slog.Info("running post-generate tasks")

			cmd := exec.Command("goimports", "-w", ".")
			cmd.Dir = out
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("run goimports: %w", err)
			}

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "out",
				Aliases:     []string{"o"},
				Usage:       "path of the output directory",
				Required:    false,
				Destination: &out,
				Value:       "./",
			},
		},
	}
}
