package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func WriteFile(t *testing.T, directory, name, content string) string {
	t.Helper()

	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

func MakeDirectory(t *testing.T, parent, name string) string {
	t.Helper()

	path := filepath.Join(parent, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}

	return path
}

func TempFile(name, content string) func(*testing.T) string {
	return func(t *testing.T) string {
		t.Helper()

		return WriteFile(t, t.TempDir(), name, content)
	}
}

func StaticPath(path string) func(*testing.T) string {
	return func(*testing.T) string { return path }
}

func NestedTree(topContent, nestedContent string) func(*testing.T) string {
	return func(t *testing.T) string {
		t.Helper()

		directory := t.TempDir()
		WriteFile(t, directory, "top.txt", topContent)
		subDirectory := MakeDirectory(t, directory, "sub")
		WriteFile(t, subDirectory, "nested.txt", nestedContent)

		return directory
	}
}
