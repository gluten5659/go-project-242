package main

import (
	"context"
	"os"

	"code/internal/cliapp"
)

func main() {
	_ = cliapp.NewCommand().Run(context.Background(), os.Args)
}
