package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/urfave/cli/v2"

	"github.com/sumup/sumup-go/codegen/pkg/builder"
)

func Generate() *cli.Command {
	var (
		out     string
		modName string
		pkgName string
		name    string
		force   bool
	)

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

			spec, err := openapi3.NewLoader().LoadFromFile(specs)
			if err != nil {
				return err
			}

			builder := builder.New(builder.Config{
				Out:     out,
				Module:  modName,
				PkgName: pkgName,
				Name:    name,
			})

			if err := builder.Load(spec); err != nil {
				return fmt.Errorf("load spec: %w", err)
			}

			if force {
				slog.Info("bootstrapping new package")

				_, err := os.Stat(path.Join(out, "go.mod"))
				newRepo := errors.Is(err, os.ErrNotExist)
				if newRepo {
					cmd := exec.Command("go", "mod", "init", modName)
					cmd.Dir = out
					if err := cmd.Run(); err != nil {
						return fmt.Errorf("init go module: %w", err)
					}
				}

				if err := builder.Bootstrap(); err != nil {
					return fmt.Errorf("boostrap package: %w", err)
				}
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
			&cli.StringFlag{
				Name:        "module",
				Aliases:     []string{"m", "mod"},
				Usage:       "name of the generated module",
				Required:    true,
				Destination: &modName,
			},
			&cli.StringFlag{
				Name:        "package",
				Aliases:     []string{"p", "pkg"},
				Usage:       "name of the generated package",
				Required:    true,
				Destination: &pkgName,
			},
			&cli.StringFlag{
				Name:        "name",
				Aliases:     []string{"n"},
				Usage:       "name of your service",
				Required:    true,
				Destination: &name,
			},
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Usage:       "force creation of all base files that can later be modified by the user",
				Destination: &force,
			},
		},
	}
}
