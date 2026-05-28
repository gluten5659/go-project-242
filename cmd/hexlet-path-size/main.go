package main

import (
	"fmt"
	"os"

	"code/internal/cliapp"
	"code/internal/humanize"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	output, path, err := cliapp.RunCli(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, cliapp.UserMessage(err, path))

		return cliapp.ExitCodeFor(err)
	}

	fmt.Println(humanize.FormatLine(output, path))

	return cliapp.ExitOK
}
