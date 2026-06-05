package cliapp

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"code"
	"code/internal/dirsize"
	"code/internal/output"

	"github.com/urfave/cli/v3"
)

var errUsage = errors.New("usage error")

const (
	exitGeneric    = 1
	exitUsage      = 64
	exitDataErr    = 65
	exitNoInput    = 66
	exitIOErr      = 74
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
				return userError(err)
			}

			line := output.FormatLine(size, path)
			if _, err := fmt.Fprintln(cmd.Root().Writer, line); err != nil {
				return cli.Exit(fmt.Errorf("write output: %w", err), exitIOErr)
			}

			return nil
		},
	}
}

func userError(err error) error {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return cli.Exit(pathMessage("path not found", err), exitNoInput)
	case errors.Is(err, fs.ErrPermission):
		return cli.Exit(pathMessage("permission denied", err), exitPermission)
	case errors.Is(err, dirsize.ErrUnsupportedPath):
		return cli.Exit(err, exitDataErr)
	default:
		return cli.Exit(err, exitGeneric)
	}
}

func pathMessage(label string, err error) string {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return fmt.Sprintf("%s: %q", label, pathErr.Path)
	}

	return fmt.Sprintf("%s: %s", label, err)
}
