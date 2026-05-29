package cliapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"code/internal/dirsize"

	"github.com/urfave/cli/v3"
)

func TestCommandOutput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		setup    func(t *testing.T) string
		flags    []string
		wantSize string
	}{
		{
			desc:     "regular file raw bytes",
			setup:    tempFile("a.txt", "hello"),
			wantSize: "5B",
		},
		{
			desc:     "human-readable format with -H",
			setup:    tempFile("big.dat", strings.Repeat("\x00", 1024*3/2)),
			flags:    []string{"-H"},
			wantSize: "1.5KB",
		},
		{
			desc: "directory non-recursive ignores nested files",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				writeTestFile(t, directory, "top.txt", "hello")
				subDir := makeSubDir(t, directory, "sub")
				writeTestFile(t, subDir, "nested.txt", "ignored")

				return directory
			},
			wantSize: "5B",
		},
		{
			desc: "directory recursive sums nested files",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				writeTestFile(t, directory, "top.txt", "hello")
				subDir := makeSubDir(t, directory, "sub")
				writeTestFile(t, subDir, "nested.txt", "world!")

				return directory
			},
			flags:    []string{"-r"},
			wantSize: "11B",
		},
		{
			desc: "hidden files excluded by default",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				writeTestFile(t, directory, "visible.txt", "hello")
				writeTestFile(t, directory, ".hidden.txt", "xx")

				return directory
			},
			wantSize: "5B",
		},
		{
			desc: "hidden files included with -a",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				writeTestFile(t, directory, "visible.txt", "hello")
				writeTestFile(t, directory, ".hidden.txt", "xx")

				return directory
			},
			flags:    []string{"-a"},
			wantSize: "7B",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			path := tC.setup(t)
			args := append([]string{}, tC.flags...)

			line, err := runCLI(t, append(args, path)...)
			if err != nil {
				t.Fatalf("runCLI returned error: %v", err)
			}

			want := fmt.Sprintf("%s\t%s", tC.wantSize, path)
			if line != want {
				t.Errorf("output = %q, want %q", line, want)
			}
		})
	}
}

func TestCommandUsageErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
		args []string
	}{
		{
			desc: "no path",
			args: []string{},
		},
		{
			desc: "two paths",
			args: []string{"/tmp/a", "/tmp/b"},
		},
		{
			desc: "unknown flag",
			args: []string{"--bogus", "/tmp/a"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			_, err := runCLI(t, tC.args...)
			if got := exitCodeOf(t, err); got != exitUsage {
				t.Errorf("exit code = %d, want %d", got, exitUsage)
			}

			if !strings.HasPrefix(err.Error(), "usage error") {
				t.Errorf("error %q does not start with %q", err.Error(), "usage error")
			}
		})
	}
}

func TestCommandReportsExitCodeForMissingPath(t *testing.T) {
	t.Parallel()

	_, err := runCLI(t, "/no/such/path")

	if got := exitCodeOf(t, err); got != exitNoInput {
		t.Errorf("exit code = %d, want %d", got, exitNoInput)
	}

	if !strings.Contains(err.Error(), "path not found") {
		t.Errorf("error %q does not mention %q", err.Error(), "path not found")
	}
}

func TestUserError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		err      error
		wantCode int
		wantMsg  string
	}{
		{
			desc:     "missing path maps to no-input",
			err:      &fs.PathError{Op: "lstat", Path: "/x", Err: fs.ErrNotExist},
			wantCode: exitNoInput,
			wantMsg:  `path not found: "/x"`,
		},
		{
			desc:     "permission denied maps to permission code",
			err:      &fs.PathError{Op: "open", Path: "/x", Err: fs.ErrPermission},
			wantCode: exitPermission,
			wantMsg:  `permission denied: "/x"`,
		},
		{
			desc:     "unsupported path maps to data error",
			err:      fmt.Errorf("%w: %q", dirsize.ErrUnsupportedPath, "/x"),
			wantCode: exitDataErr,
			wantMsg:  `unsupported path type: "/x"`,
		},
		{
			desc:     "unknown error falls back to generic",
			err:      errors.New("boom"),
			wantCode: exitGeneric,
			wantMsg:  "boom",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			got := userError(tC.err)
			if code := exitCodeOf(t, got); code != tC.wantCode {
				t.Errorf("exit code = %d, want %d", code, tC.wantCode)
			}

			if got.Error() != tC.wantMsg {
				t.Errorf("message = %q, want %q", got.Error(), tC.wantMsg)
			}
		})
	}
}

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var stdout, stderr bytes.Buffer

	cmd := NewCommand()
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr
	cmd.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}

	err := cmd.Run(context.Background(), append([]string{"hexlet-path-size"}, args...))

	return strings.TrimRight(stdout.String(), "\n"), err
}

func exitCodeOf(t *testing.T, err error) int {
	t.Helper()

	var coder cli.ExitCoder
	if !errors.As(err, &coder) {
		t.Fatalf("error %v does not implement cli.ExitCoder", err)
	}

	return coder.ExitCode()
}

func tempFile(name, content string) func(*testing.T) string {
	return func(t *testing.T) string {
		t.Helper()

		return writeTestFile(t, t.TempDir(), name, content)
	}
}

func writeTestFile(t *testing.T, directory, name, content string) string {
	t.Helper()

	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

func makeSubDir(t *testing.T, parent, name string) string {
	t.Helper()

	path := filepath.Join(parent, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}

	return path
}
