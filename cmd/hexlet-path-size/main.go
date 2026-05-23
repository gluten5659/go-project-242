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
		fmt.Fprintln(os.Stderr, cliapp.UserMessage(err, path))
		os.Exit(cliapp.ExitCodeFor(err))
	}
	fmt.Println(humanize.FormatLine(output, path))
}
