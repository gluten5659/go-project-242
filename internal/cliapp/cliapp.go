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

var ErrUsage = errors.New("usage error")

const (
	ExitOK         = 0
	ExitGeneric    = 1
	ExitUsage      = 64
	ExitDataErr    = 65
	ExitNoInput    = 66
	ExitPermission = 77
)

func ExitCodeFor(err error) int {
	switch {
	case err == nil:
		return ExitOK
	case errors.Is(err, ErrUsage):
		return ExitUsage
	case errors.Is(err, scan.ErrPathNotFound):
		return ExitNoInput
	case errors.Is(err, scan.ErrPermissionDenied):
		return ExitPermission
	case errors.Is(err, scan.ErrUnsupportedPath):
		return ExitDataErr
	default:
		return ExitGeneric
	}
}

func RunCli(args []string) (string, error) {
	var (
		formatNeeded  bool
		includeHidden bool
		recursive     bool
		line          string
	)

	cmd := &cli.Command{
		Name:      "hexlet-path-size",
		Usage:     "print size of a file or directory",
		ArgsUsage: "<path>",
		OnUsageError: func(_ context.Context, _ *cli.Command, usageErr error, _ bool) error {
			return fmt.Errorf("%w: %s", ErrUsage, usageErr.Error())
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
				return fmt.Errorf("%w: exactly one file path is required", ErrUsage)
			}

			path := cmd.Args().Get(0)

			size, err := code.GetPathSize(path, recursive, formatNeeded, includeHidden)
			if err != nil {
				return err
			}

			line = humanize.FormatLine(size, path)

			return nil
		},
	}

	if err := cmd.Run(context.Background(), args); err != nil {
		return "", err
	}

	return line, nil
}
