package cliapp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"code/internal/scan"
)

func TestRunCli(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc       string
		setup      func(t *testing.T) string
		flags      []string
		wantOutput string
		wantErrIs  error
		wantErr    bool
	}{
		{
			desc:       "regular file raw bytes",
			setup:      tempFile("a.txt", "hello"),
			wantOutput: "5B",
		},
		{
			desc:       "human-readable format with -H",
			setup:      tempFile("big.dat", strings.Repeat("\x00", 1024*3/2)),
			flags:      []string{"-H"},
			wantOutput: "1.5KB",
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
			wantOutput: "5B",
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
			flags:      []string{"-r"},
			wantOutput: "11B",
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
			wantOutput: "5B",
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
			flags:      []string{"-a"},
			wantOutput: "7B",
		},
		{
			desc:      "nonexistent path returns error",
			setup:     staticPath("/no/such/path"),
			wantErr:   true,
			wantErrIs: scan.ErrPathNotFound,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()
			path := tC.setup(t)
			args := append([]string{"hexlet-path-size"}, tC.flags...)
			args = append(args, path)

			output, gotPath, err := RunCli(args)
			if (err != nil) != tC.wantErr {
				t.Fatalf("RunCli error = %v, wantErr %v", err, tC.wantErr)
			}

			if tC.wantErrIs != nil && !errors.Is(err, tC.wantErrIs) {
				t.Errorf("RunCli error = %v, want errors.Is(_, %v)", err, tC.wantErrIs)
			}

			if tC.wantErr {
				return
			}

			if output != tC.wantOutput {
				t.Errorf("RunCli output = %q, want %q", output, tC.wantOutput)
			}

			if gotPath != path {
				t.Errorf("RunCli path = %q, want %q", gotPath, path)
			}
		})
	}
}

func TestRunCliArgs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
		args []string
	}{
		{
			desc: "no path returns error",
			args: []string{"hexlet-path-size"},
		},
		{
			desc: "two paths returns error",
			args: []string{"hexlet-path-size", "/tmp/a", "/tmp/b"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			_, _, err := RunCli(tC.args)
			if !errors.Is(err, ErrUsage) {
				t.Fatalf("RunCli error = %v, want errors.Is(_, ErrUsage)", err)
			}
		})
	}
}

func TestUserMessage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
		err  error
		path string
		want string
	}{
		{"no error", nil, "/x", ""},
		{
			desc: "usage error keeps its own message",
			err:  fmt.Errorf("%w: exactly one file path is required", ErrUsage),
			path: "",
			want: "usage error: exactly one file path is required",
		},
		{
			desc: "path not found",
			err:  fmt.Errorf("%w: %q: underlying", scan.ErrPathNotFound, "/x"),
			path: "/x",
			want: `path not found: "/x"`,
		},
		{
			desc: "permission denied",
			err:  fmt.Errorf("%w: %q: underlying", scan.ErrPermissionDenied, "/x"),
			path: "/x",
			want: `permission denied: "/x"`,
		},
		{
			desc: "unsupported file type",
			err:  fmt.Errorf("%w: %q", scan.ErrUnsupportedPath, "/x"),
			path: "/x",
			want: `unsupported file type: "/x"`,
		},
		{
			desc: "read failed",
			err:  fmt.Errorf("%w: %q: underlying", scan.ErrReadFailed, "/x"),
			path: "/x",
			want: `cannot read: "/x"`,
		},
		{
			desc: "unknown error falls back to its own message",
			err:  errors.New("boom"),
			path: "/x",
			want: "boom",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			got := UserMessage(tC.err, tC.path)
			if got != tC.want {
				t.Errorf("UserMessage(%v, %q) = %q, want %q", tC.err, tC.path, got, tC.want)
			}
		})
	}
}

func TestExitCodeFor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
		err  error
		want int
	}{
		{"no error", nil, ExitOK},
		{"usage error", ErrUsage, ExitUsage},
		{"path not found", scan.ErrPathNotFound, ExitNoInput},
		{"permission denied", scan.ErrPermissionDenied, ExitPermission},
		{"unsupported path", scan.ErrUnsupportedPath, ExitDataErr},
		{"unknown error", errors.New("boom"), ExitGeneric},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			got := ExitCodeFor(tC.err)
			if got != tC.want {
				t.Errorf("ExitCodeFor(%v) = %d, want %d", tC.err, got, tC.want)
			}
		})
	}
}

func tempFile(name, content string) func(*testing.T) string {
	return func(t *testing.T) string {
		t.Helper()

		return writeTestFile(t, t.TempDir(), name, content)
	}
}

func staticPath(path string) func(*testing.T) string {
	return func(*testing.T) string { return path }
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
