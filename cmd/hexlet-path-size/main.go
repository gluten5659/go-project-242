package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"code"

	"github.com/urfave/cli/v3"
)

var ErrUsage = errors.New("usage error")

const (
	exitOK         = 0
	exitGeneric    = 1
	exitUsage      = 64
	exitDataErr    = 65
	exitNoInput    = 66
	exitPermission = 77
)

func exitCodeFor(err error) int {
	switch {
	case err == nil:
		return exitOK
	case errors.Is(err, ErrUsage):
		return exitUsage
	case errors.Is(err, code.ErrPathNotFound):
		return exitNoInput
	case errors.Is(err, code.ErrPermissionDenied):
		return exitPermission
	case errors.Is(err, code.ErrUnsupportedPath):
		return exitDataErr
	default:
		return exitGeneric
	}
}

func main() {
	output, path, err := runCli(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCodeFor(err))
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
				return fmt.Errorf("%w: exactly one file path is required", ErrUsage)
			}
			path = cmd.Args().Get(0)
			result, err = code.GetPathSize(path, recursive, formatNeeded, listHidden)
			return err
		},
	}
	err = cmd.Run(context.Background(), args)
	return result, path, err
}
