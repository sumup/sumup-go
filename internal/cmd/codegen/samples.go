package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func Samples() *cli.Command {
	return &cli.Command{
		Name:  "samples",
		Usage: "Generate operation samples",
		Args:  true,
		Action: func(c *cli.Context) error {
			if !c.Args().Present() {
				return fmt.Errorf("empty argument, path to openapi specs expected")
			}

			b, err := loadBuilder(c.Args().First(), "./")
			if err != nil {
				return err
			}

			samples, err := b.BuildSamples()
			if err != nil {
				return fmt.Errorf("build samples: %w", err)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(samples); err != nil {
				return fmt.Errorf("encode samples: %w", err)
			}

			return nil
		},
	}
}
