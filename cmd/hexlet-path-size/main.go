package main

import (
	"fmt"
	"os"

	"code/internal/cliapp"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	line, err := cliapp.RunCli(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return cliapp.ExitCodeFor(err)
	}

	fmt.Println(line)

	return cliapp.ExitOK
}
