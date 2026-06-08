package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
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

			if err := os.MkdirAll(out, os.ModePerm); err != nil {
				return fmt.Errorf("create output directory %q: %w", out, err)
			}

			builder, err := loadBuilder(c.Args().First(), out)
			if err != nil {
				return err
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
