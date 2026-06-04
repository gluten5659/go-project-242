package main

import (
	"context"
	"fmt"
	"os"

	"code/internal/cliapp"
)

func main() {
	if err := cliapp.NewCommand().Run(context.Background(), os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}
