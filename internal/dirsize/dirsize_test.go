package dirsize

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"code/internal/testutil"
)

type measureCase struct {
	desc          string
	setup         func(t *testing.T) string
	includeHidden bool
	recursive     bool
	want          int64
	wantErrIs     error
}

func runMeasureCases(t *testing.T, testCases []measureCase) {
	t.Helper()

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			path := tC.setup(t)

			got, err := Measure(path, tC.includeHidden, tC.recursive)
			if tC.wantErrIs != nil {
				if !errors.Is(err, tC.wantErrIs) {
					t.Fatalf("Measure error = %v, want errors.Is(_, %v)", err, tC.wantErrIs)
				}

				return
			}

			if err != nil {
				t.Fatalf("Measure unexpected error: %v", err)
			}

			if got != tC.want {
				t.Errorf("Measure = %d, want %d", got, tC.want)
			}
		})
	}
}

func TestMeasureFile(t *testing.T) {
	t.Parallel()

	runMeasureCases(t, []measureCase{
		{
			desc:  "file with content",
			setup: testutil.TempFile("data.txt", "hello"),
			want:  5,
		},
		{
			desc:  "empty file",
			setup: testutil.TempFile("empty.txt", ""),
			want:  0,
		},
	})
}

func TestMeasureDirectory(t *testing.T) {
	t.Parallel()

	runMeasureCases(t, []measureCase{
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
				testutil.WriteFile(t, directory, "a.txt", "hello")

				return directory
			},
			want: 5,
		},
		{
			desc: "multiple files summed",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				testutil.WriteFile(t, directory, "a.txt", "hello")
				testutil.WriteFile(t, directory, "b.txt", "world!")

				return directory
			},
			want: 11,
		},
	})
}

func TestMeasureHidden(t *testing.T) {
	t.Parallel()

	runMeasureCases(t, []measureCase{
		{
			desc:  "hidden file passed directly is always counted",
			setup: testutil.TempFile(".secret.txt", "shh"),
			want:  3,
		},
		{
			desc:  "hidden entry excluded from a walk by default",
			setup: hiddenTree(),
			want:  5,
		},
		{
			desc:          "hidden entry included with includeHidden",
			setup:         hiddenTree(),
			includeHidden: true,
			want:          7,
		},
	})
}

func TestMeasureRecursive(t *testing.T) {
	t.Parallel()

	runMeasureCases(t, []measureCase{
		{
			desc:      "nested folder skipped without recursive",
			setup:     testutil.NestedTree("hello", "ignored"),
			recursive: false,
			want:      5,
		},
		{
			desc:      "nested folder summed with recursive",
			setup:     testutil.NestedTree("hello", "world!"),
			recursive: true,
			want:      11,
		},
	})
}

func TestMeasureSymlink(t *testing.T) {
	t.Parallel()

	runMeasureCases(t, []measureCase{
		{
			desc: "symlink reports size of link entry, not target",
			setup: func(t *testing.T) string {
				t.Helper()
				skipWithoutSymlinks(t)
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
			desc: "folder sums link entry size, not target",
			setup: func(t *testing.T) string {
				t.Helper()
				skipWithoutSymlinks(t)
				directory := t.TempDir()
				testutil.WriteFile(t, directory, "a.txt", "hello")

				if err := os.Symlink("xy", filepath.Join(directory, "link")); err != nil {
					t.Fatal(err)
				}

				return directory
			},
			want: 5 + int64(len("xy")),
		},
	})
}

func TestMeasureErrors(t *testing.T) {
	t.Parallel()

	runMeasureCases(t, []measureCase{
		{
			desc:      "nonexistent path",
			setup:     testutil.StaticPath("/definitely/not/exists/here"),
			wantErrIs: fs.ErrNotExist,
		},
		{
			desc: "FIFO is an unsupported path type",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()

				fifoPath := filepath.Join(directory, "pipe")
				if err := syscall.Mkfifo(fifoPath, 0o644); err != nil {
					t.Fatal(err)
				}

				return fifoPath
			},
			wantErrIs: ErrUnsupportedPath,
		},
		{
			desc: "unreadable nested folder breaks recursive walk",
			setup: func(t *testing.T) string {
				t.Helper()
				directory := t.TempDir()
				subDir := testutil.MakeDirectory(t, directory, "locked")
				testutil.WriteFile(t, subDir, "inside.txt", "secret")

				if err := os.Chmod(subDir, 0o600); err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() {
					_ = os.Chmod(subDir, 0o700)
				})

				return directory
			},
			recursive: true,
			wantErrIs: fs.ErrPermission,
		},
	})
}

func hiddenTree() func(*testing.T) string {
	return func(t *testing.T) string {
		t.Helper()
		directory := t.TempDir()
		testutil.WriteFile(t, directory, "visible.txt", "hello")
		testutil.WriteFile(t, directory, ".hidden.txt", "xx")

		return directory
	}
}

func skipWithoutSymlinks(t *testing.T) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("symlinks require special privileges on Windows")
	}
}
