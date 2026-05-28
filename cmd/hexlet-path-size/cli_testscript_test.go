package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"hexlet-path-size": main,
	})
}

func TestCLI(t *testing.T) {
	t.Parallel()

	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
		Condition: func(cond string) (bool, error) {
			if cond == "root" {
				return os.Geteuid() == 0, nil
			}

			return false, fmt.Errorf("unknown condition: %s", cond)
		},
	})
}
