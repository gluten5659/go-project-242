package cliapp

import (
	"context"
	"errors"
	"fmt"

	"code"
	"code/internal/humanize"
	"code/internal/scan"

	"github.com/urfave/cli/v3"
)

var errUsage = errors.New("usage error")

const (
	exitGeneric    = 1
	exitUsage      = 64
	exitDataErr    = 65
	exitNoInput    = 66
	exitPermission = 77
)

func NewCommand() *cli.Command {
	var (
		formatNeeded  bool
		includeHidden bool
		recursive     bool
	)

	return &cli.Command{
		Name:      "hexlet-path-size",
		Usage:     "print size of a file or directory",
		ArgsUsage: "<path>",
		OnUsageError: func(_ context.Context, _ *cli.Command, usageErr error, _ bool) error {
			return cli.Exit(fmt.Errorf("%w: %s", errUsage, usageErr.Error()), exitUsage)
		},
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
				Destination: &includeHidden,
			},
			&cli.BoolFlag{
				Name:        "recursive",
				Usage:       "Recursive sizes",
				Aliases:     []string{"r"},
				Destination: &recursive,
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() != 1 {
				return cli.Exit(fmt.Errorf("%w: exactly one file path is required", errUsage), exitUsage)
			}

			path := cmd.Args().Get(0)

			size, err := code.GetPathSize(path, recursive, formatNeeded, includeHidden)
			if err != nil {
				return cli.Exit(err, exitCodeFor(err))
			}

			_, _ = fmt.Fprintln(cmd.Root().Writer, humanize.FormatLine(size, path))

			return nil
		},
	}
}

func exitCodeFor(err error) int {
	switch {
	case errors.Is(err, scan.ErrPathNotFound):
		return exitNoInput
	case errors.Is(err, scan.ErrPermissionDenied):
		return exitPermission
	case errors.Is(err, scan.ErrUnsupportedPath):
		return exitDataErr
	default:
		return exitGeneric
	}
}
