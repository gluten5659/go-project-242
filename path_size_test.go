package code

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatSize(t *testing.T) {
	testCases := []struct {
		desc         string
		byteCount    int64
		formatNeeded bool
		want         string
	}{
		{"raw bytes when format disabled", 1500, false, "1500B"},
		{"under 1KB stays as bytes", 500, true, "500B"},
		{"exactly 1023 bytes stays as bytes", 1024 - 1, true, "1023B"},
		{"exactly 1KB", 1024, true, "1.0KB"},
		{"1.5 MB", 1024 * 1024 * 3 / 2, true, "1.5MB"},
		{"2 GB", 1024 * 1024 * 1024 * 2, true, "2.0GB"},
		{"zero", 0, true, "0B"},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := formatOutput(tC.byteCount, tC.formatNeeded)
			if got != tC.want {
				t.Errorf("formatOutput(%d, %v) = %q, want %q",
					tC.byteCount, tC.formatNeeded, got, tC.want)
			}
		})
	}
}

func TestPickUnit(t *testing.T) {
	testCases := []struct {
		desc      string
		byteCount int64
		wantValue float64
		wantUnit  string
	}{
		{"zero", 0, 0, "B"},
		{"1023", 1024 - 1, 1023, "B"},
		{"1024", 1024, 1, "KB"},
		{"1.5 KB", 1024 * 3 / 2, 1.5, "KB"},
		{"1 EB", 1024 * 1024 * 1024 * 1024 * 1024 * 1024, 1, "EB"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gotValue, gotUnit := pickUnit(tC.byteCount)
			if gotValue != tC.wantValue || gotUnit != tC.wantUnit {
				t.Errorf("pickUnit(%d) = (%v, %q), want (%v, %q)",
					tC.byteCount, gotValue, gotUnit,
					tC.wantValue, tC.wantUnit)
			}
		})
	}
}

func TestGetSize(t *testing.T) {
	testCases := []struct {
		desc    string
		setup   func(t *testing.T) string
		want    int64
		wantErr bool
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
				directory := t.TempDir()
				writeTestFile(t, directory, "a.txt", "12345")
				return directory
			},
			want: 5,
		},
		{
			desc:    "nonexistent path",
			setup:   staticPath("/definitely/not/exists/here"),
			wantErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			path := tC.setup(t)
			got, err := getSize(path, false, false)
			if (err != nil) != tC.wantErr {
				t.Fatalf("getSize error = %v, wantErr %v", err, tC.wantErr)
			}
			if got != tC.want {
				t.Errorf("getSize = %d, want %d", got, tC.want)
			}
		})
	}
}

func TestGetFolderSize(t *testing.T) {
	testCases := []struct {
		desc       string
		setup      func(t *testing.T) string
		listHidden bool
		recursive  bool
		want       int64
		wantErr    bool
	}{
		{
			desc:  "empty folder",
			setup: func(t *testing.T) string { return t.TempDir() },
			want:  0,
		},
		{
			desc: "single file",
			setup: func(t *testing.T) string {
				directory := t.TempDir()
				writeTestFile(t, directory, "a.txt", "hello")
				return directory
			},
			want: 5,
		},
		{
			desc: "multiple files summed",
			setup: func(t *testing.T) string {
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
			desc:    "nonexistent folder",
			setup:   staticPath("/nope/nada/nothing"),
			wantErr: true,
		},
		{
			desc: "nested folder with 0600 mode breaks recursive walk",
			setup: func(t *testing.T) string {
				directory := t.TempDir()
				subDir := makeSubDir(t, directory, "locked")
				writeTestFile(t, subDir, "inside.txt", "secret")
				if err := os.Chmod(subDir, 0600); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					_ = os.Chmod(subDir, 0700)
				})
				return directory
			},
			recursive: true,
			wantErr:   true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			folderPath := tC.setup(t)
			got, err := getFolderSize(folderPath, tC.listHidden, tC.recursive)
			if (err != nil) != tC.wantErr {
				t.Fatalf("getFolderSize error = %v, wantErr %v", err, tC.wantErr)
			}
			if got != tC.want {
				t.Errorf("getFolderSize = %d, want %d", got, tC.want)
			}
		})
	}
}

func TestGetPathSize(t *testing.T) {
	testCases := []struct {
		desc         string
		setup        func(t *testing.T) string
		recursive    bool
		formatNeeded bool
		listHidden   bool
		want         string
		wantErr      bool
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
			desc:    "nonexistent path returns error",
			setup:   staticPath("/no/such/path"),
			wantErr: true,
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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func makeSubDir(t *testing.T, parent, name string) string {
	t.Helper()
	path := filepath.Join(parent, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	return path
}
