package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCli(t *testing.T) {
	testCases := []struct {
		desc       string
		setup      func(t *testing.T) string
		flags      []string
		wantOutput string
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
				directory := t.TempDir()
				writeTestFile(t, directory, "visible.txt", "hello")
				writeTestFile(t, directory, ".hidden.txt", "xx")
				return directory
			},
			flags:      []string{"-a"},
			wantOutput: "7B",
		},
		{
			desc:    "nonexistent path returns error",
			setup:   staticPath("/no/such/path"),
			wantErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			path := tC.setup(t)
			args := append([]string{"hexlet-path-size"}, tC.flags...)
			args = append(args, path)

			output, gotPath, err := runCli(args)
			if (err != nil) != tC.wantErr {
				t.Fatalf("runCli error = %v, wantErr %v", err, tC.wantErr)
			}
			if tC.wantErr {
				return
			}
			if output != tC.wantOutput {
				t.Errorf("runCli output = %q, want %q", output, tC.wantOutput)
			}
			if gotPath != path {
				t.Errorf("runCli path = %q, want %q", gotPath, path)
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
