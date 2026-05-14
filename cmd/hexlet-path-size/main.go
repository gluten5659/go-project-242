package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"code"

	"github.com/urfave/cli/v3"
)

func main() {
	output, path, err := runCli(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s\t%s\n", output, path)
}

func runCli(args []string) (string, string, error) {
	var (
		formatNeeded bool
		listHidden   bool
		recursive    bool
		result       string
		path         string
		err          error
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
			if cmd.Args().Len() != 1 {
				return errors.New("exactly one file path is required")
			}
			path = cmd.Args().Get(0)
			result, err = code.GetPathSize(path, recursive, formatNeeded, listHidden)
			return err
		},
	}
	err = cmd.Run(context.Background(), args)
	return result, path, err
}
