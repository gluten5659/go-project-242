package code

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"code/internal/scan"
)

func TestGetPathSize(t *testing.T) {
	testCases := []struct {
		desc         string
		setup        func(t *testing.T) string
		recursive    bool
		formatNeeded bool
		listHidden   bool
		want         string
		wantErr      bool
		wantErrIs    error
	}{
		{
			desc:         "raw bytes for file",
			setup:        tempFile("a.txt", "hello"),
			formatNeeded: false,
			want:         "5B",
		},
		{
			desc:         "formatted KB for file",
			setup:        tempFile("big.dat", strings.Repeat("\x00", 1024*3/2)),
			formatNeeded: true,
			want:         "1.5KB",
		},
		{
			desc:         "recursive directory total",
			setup:        nestedTree("hello", "world!"),
			recursive:    true,
			formatNeeded: false,
			want:         "11B",
		},
		{
			desc:      "nonexistent path returns error",
			setup:     staticPath("/no/such/path"),
			wantErr:   true,
			wantErrIs: scan.ErrPathNotFound,
		},
		{
			desc:         "hidden file path is shown despite listHidden being false",
			setup:        tempFile(".env", "PORT=8080"),
			formatNeeded: false,
			listHidden:   false,
			want:         "9B",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			path := tC.setup(t)
			got, err := GetPathSize(path, tC.recursive, tC.formatNeeded, tC.listHidden)
			if (err != nil) != tC.wantErr {
				t.Fatalf("GetPathSize error = %v, wantErr %v", err, tC.wantErr)
			}
			if tC.wantErrIs != nil && !errors.Is(err, tC.wantErrIs) {
				t.Errorf("GetPathSize error = %v, want errors.Is(_, %v)", err, tC.wantErrIs)
			}
			if got != tC.want {
				t.Errorf("GetPathSize = %q, want %q", got, tC.want)
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

func nestedTree(topContent, nestedContent string) func(*testing.T) string {
	return func(t *testing.T) string {
		t.Helper()
		directory := t.TempDir()
		writeTestFile(t, directory, "top.txt", topContent)
		subDir := makeSubDir(t, directory, "sub")
		writeTestFile(t, subDir, "nested.txt", nestedContent)
		return directory
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
