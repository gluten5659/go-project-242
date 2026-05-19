package scan

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestSize(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc      string
		setup     func(t *testing.T) string
		want      int64
		wantErr   bool
		wantErrIs error
	}{
		{
			desc:  "regular file with content",
			setup: tempFile("data.txt", "hello"),
			want:  5,
		},
		{
			desc:  "empty file",
			setup: tempFile("empty.txt", ""),
			want:  0,
		},
		{
			desc:  "hidden file passed directly is counted even when listHidden is false",
			setup: tempFile(".secret.txt", "shh"),
			want:  3,
		},
		{
			desc: "directory delegates to getFolderSize",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				writeTestFile(t, directory, "a.txt", "12345")
				return directory
			},
			want: 5,
		},
		{
			desc:      "nonexistent path",
			setup:     staticPath("/definitely/not/exists/here"),
			wantErr:   true,
			wantErrIs: ErrPathNotFound,
		},
		{
			desc: "symlink reports size of link entry, not target",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				linkPath := filepath.Join(directory, "link")
				if err := os.Symlink("known-target", linkPath); err != nil {
					t.Fatal(err)
				}
				return linkPath
			},
			want: int64(len("known-target")),
		},
		{
			desc: "FIFO returns ErrUnsupportedPath",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				fifoPath := filepath.Join(directory, "pipe")
				if err := syscall.Mkfifo(fifoPath, 0o644); err != nil {
					t.Fatal(err)
				}
				return fifoPath
			},
			wantErr:   true,
			wantErrIs: ErrUnsupportedPath,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()
			path := tC.setup(t)
			got, err := Size(path, false, false)
			if (err != nil) != tC.wantErr {
				t.Fatalf("Size error = %v, wantErr %v", err, tC.wantErr)
			}
			if tC.wantErrIs != nil && !errors.Is(err, tC.wantErrIs) {
				t.Errorf("Size error = %v, want errors.Is(_, %v)", err, tC.wantErrIs)
			}
			if got != tC.want {
				t.Errorf("Size = %d, want %d", got, tC.want)
			}
		})
	}
}

func TestSizeFolder(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc       string
		setup      func(t *testing.T) string
		listHidden bool
		recursive  bool
		want       int64
		wantErr    bool
		wantErrIs  error
	}{
		{
			desc: "empty folder",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			want: 0,
		},
		{
			desc: "single file",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				writeTestFile(t, directory, "a.txt", "hello")
				return directory
			},
			want: 5,
		},
		{
			desc: "multiple files summed",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				writeTestFile(t, directory, "a.txt", "hello")
				writeTestFile(t, directory, "b.txt", "world!")
				return directory
			},
			want: 11,
		},
		{
			desc: "hidden file excluded by default",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				writeTestFile(t, directory, "visible.txt", "hello")
				writeTestFile(t, directory, ".hidden.txt", "xx")
				return directory
			},
			listHidden: false,
			want:       5,
		},
		{
			desc: "hidden file included when listHidden is true",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				writeTestFile(t, directory, "visible.txt", "hello")
				writeTestFile(t, directory, ".hidden.txt", "xx")
				return directory
			},
			listHidden: true,
			want:       7,
		},
		{
			desc:      "non-recursive skips nested folder",
			setup:     nestedTree("hello", "ignored"),
			recursive: false,
			want:      5,
		},
		{
			desc:      "recursive includes nested folder",
			setup:     nestedTree("hello", "world!"),
			recursive: true,
			want:      11,
		},
		{
			desc:      "nonexistent folder",
			setup:     staticPath("/nope/nada/nothing"),
			wantErr:   true,
			wantErrIs: ErrPathNotFound,
		},
		{
			desc: "folder with symlink sums link entry size, not target",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				writeTestFile(t, directory, "a.txt", "hello")
				if err := os.Symlink("xy", filepath.Join(directory, "link")); err != nil {
					t.Fatal(err)
				}
				return directory
			},
			want: 5 + int64(len("xy")),
		},
		{
			desc: "nested folder with 0600 mode breaks recursive walk",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				subDir := makeSubDir(t, directory, "locked")
				writeTestFile(t, subDir, "inside.txt", "secret")
				if err := os.Chmod(subDir, 0o600); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					_ = os.Chmod(subDir, 0o700)
				})
				return directory
			},
			recursive: true,
			wantErr:   true,
			wantErrIs: ErrPermissionDenied,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()
			folderPath := tC.setup(t)
			got, err := Size(folderPath, tC.listHidden, tC.recursive)
			if (err != nil) != tC.wantErr {
				t.Fatalf("Size error = %v, wantErr %v", err, tC.wantErr)
			}
			if tC.wantErrIs != nil && !errors.Is(err, tC.wantErrIs) {
				t.Errorf("Size error = %v, want errors.Is(_, %v)", err, tC.wantErrIs)
			}
			if got != tC.want {
				t.Errorf("Size = %d, want %d", got, tC.want)
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
