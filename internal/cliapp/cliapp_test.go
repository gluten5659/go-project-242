package cliapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"code/internal/dirsize"
	"code/internal/testutil"

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
			setup:    testutil.TempFile("a.txt", "hello"),
			wantSize: "5B",
		},
		{
			desc:     "human-readable format with -H",
			setup:    testutil.TempFile("big.dat", strings.Repeat("\x00", 1024*3/2)),
			flags:    []string{"-H"},
			wantSize: "1.5KB",
		},
		{
			desc: "directory non-recursive ignores nested files",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				testutil.WriteFile(t, directory, "top.txt", "hello")
				subDir := testutil.MakeDirectory(t, directory, "sub")
				testutil.WriteFile(t, subDir, "nested.txt", "ignored")

				return directory
			},
			wantSize: "5B",
		},
		{
			desc: "directory recursive sums nested files",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				testutil.WriteFile(t, directory, "top.txt", "hello")
				subDir := testutil.MakeDirectory(t, directory, "sub")
				testutil.WriteFile(t, subDir, "nested.txt", "world!")

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
				testutil.WriteFile(t, directory, "visible.txt", "hello")
				testutil.WriteFile(t, directory, ".hidden.txt", "xx")

				return directory
			},
			wantSize: "5B",
		},
		{
			desc: "hidden files included with -a",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				testutil.WriteFile(t, directory, "visible.txt", "hello")
				testutil.WriteFile(t, directory, ".hidden.txt", "xx")

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

func TestCommandReportsExitCodesForErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc        string
		setup       func(t *testing.T) string
		wantCode    int
		wantMessage string
	}{
		{
			desc:        "missing path maps to no-input code",
			setup:       testutil.StaticPath("/no/such/path"),
			wantCode:    exitNoInput,
			wantMessage: "path not found",
		},
		{
			desc:        "unsupported path type maps to data-error code",
			setup:       fifoPath(),
			wantCode:    exitDataErr,
			wantMessage: "unsupported path type",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			_, err := runCLI(t, tC.setup(t))

			if got := exitCodeOf(t, err); got != tC.wantCode {
				t.Errorf("exit code = %d, want %d", got, tC.wantCode)
			}

			if !strings.Contains(err.Error(), tC.wantMessage) {
				t.Errorf("error %q does not contain %q", err.Error(), tC.wantMessage)
			}
		})
	}
}

func TestCommandReportsFailingChildPathNotRoot(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc       string
		lockedPath string
	}{
		{
			desc:       "direct child directory",
			lockedPath: "locked",
		},
		{
			desc:       "deeply nested directory",
			lockedPath: filepath.Join("a", "b", "locked"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			lockedDir := testutil.MakeDirectory(t, root, tC.lockedPath)
			testutil.WriteFile(t, lockedDir, "inside.txt", "secret")

			if err := os.Chmod(lockedDir, 0o000); err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				_ = os.Chmod(lockedDir, 0o700)
			})

			_, err := runCLI(t, "-r", root)

			if got := exitCodeOf(t, err); got != exitPermission {
				t.Errorf("exit code = %d, want %d", got, exitPermission)
			}

			if !strings.Contains(err.Error(), lockedDir) {
				t.Errorf("error %q lost the failing child path %q and fell back to a shallower path", err.Error(), lockedDir)
			}
		})
	}
}

func TestCommandReportsWriteFailure(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc        string
		writeError  error
		wantCode    int
		wantMessage string
	}{
		{
			desc:        "full output stream",
			writeError:  errors.New("no space left on device"),
			wantCode:    exitIOErr,
			wantMessage: "write output: no space left on device",
		},
		{
			desc:        "closed pipe",
			writeError:  errors.New("broken pipe"),
			wantCode:    exitIOErr,
			wantMessage: "write output: broken pipe",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			path := testutil.TempFile("a.txt", "hello")(t)

			err := runCommand(t, failingWriter{err: tC.writeError}, path)

			if got := exitCodeOf(t, err); got != tC.wantCode {
				t.Errorf("exit code = %d, want %d", got, tC.wantCode)
			}

			if err.Error() != tC.wantMessage {
				t.Errorf("message = %q, want %q", err.Error(), tC.wantMessage)
			}
		})
	}
}

type failingWriter struct {
	err error
}

func (writer failingWriter) Write([]byte) (int, error) {
	return 0, writer.err
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

	var stdout bytes.Buffer

	err := runCommand(t, &stdout, args...)

	return strings.TrimRight(stdout.String(), "\n"), err
}

func runCommand(t *testing.T, stdout io.Writer, args ...string) error {
	t.Helper()

	cmd := NewCommand()
	cmd.Writer = stdout
	cmd.ErrWriter = &bytes.Buffer{}
	cmd.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}

	return cmd.Run(context.Background(), append([]string{"hexlet-path-size"}, args...))
}

func exitCodeOf(t *testing.T, err error) int {
	t.Helper()

	var coder cli.ExitCoder
	if !errors.As(err, &coder) {
		t.Fatalf("error %v does not implement cli.ExitCoder", err)
	}

	return coder.ExitCode()
}

func fifoPath() func(*testing.T) string {
	return func(t *testing.T) string {
		t.Helper()

		path := filepath.Join(t.TempDir(), "pipe")
		if err := syscall.Mkfifo(path, 0o644); err != nil {
			t.Fatal(err)
		}

		return path
	}
}
