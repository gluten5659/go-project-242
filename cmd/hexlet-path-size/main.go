package main

import (
	"context"
	"fmt"
	"log"
	"os"

	code "code"

	"github.com/urfave/cli/v3"
)

func main() {
	var (
		formatNeeded bool
		listHidden   bool
		recursive    bool
	)

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "human",
				Usage:       "Converts B into more readable KB/MB/GB etc.",
				Aliases:     []string{"H"},
				Destination: &formatNeeded,
			},
			&cli.BoolFlag{
				Name:        "all",
				Usage:       "Allow hidden files",
				Aliases:     []string{"a"},
				Destination: &listHidden,
			},
			&cli.BoolFlag{
				Name:        "recursive",
				Usage:       "Recursive sizes",
				Aliases:     []string{"r"},
				Destination: &recursive,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			path := cmd.Args().Get(0)
			result, err := code.GetPathSize(path, formatNeeded, listHidden, recursive)
			if err != nil {
				return err
			}
			fmt.Printf("%s\t%s\n", result, path)
			return nil
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
