package main

import (
	"fmt"
	"os"

	"code/internal/cliapp"
	"code/internal/humanize"
)

func main() {
	output, path, err := cliapp.RunCli(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cliapp.ExitCodeFor(err))
	}
	fmt.Println(humanize.Line(output, path))
}
