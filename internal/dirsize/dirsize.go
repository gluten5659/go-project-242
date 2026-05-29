package dirsize

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrUnsupportedPath = errors.New("unsupported path type")

func Measure(path string, includeHidden, recursive bool) (int64, error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}

	mode := stat.Mode()
	switch {
	case mode.IsDir():
		return measureFolder(path, includeHidden, recursive)
	case mode.IsRegular(), mode&os.ModeSymlink != 0:
		return stat.Size(), nil
	default:
		return 0, fmt.Errorf("%w: %q", ErrUnsupportedPath, path)
	}
}

func measureFolder(folderPath string, includeHidden bool, recursive bool) (int64, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return 0, err
	}

	var folderSize int64

	for _, entry := range entries {
		if shouldSkipEntry(entry, includeHidden, recursive) {
			continue
		}

		childPath := filepath.Join(folderPath, entry.Name())

		size, err := Measure(childPath, includeHidden, recursive)
		if err != nil {
			return 0, err
		}

		folderSize += size
	}

	return folderSize, nil
}

func shouldSkipEntry(entry os.DirEntry, includeHidden, recursive bool) bool {
	if !includeHidden && isHiddenName(entry.Name()) {
		return true
	}

	if !recursive && entry.IsDir() {
		return true
	}

	return false
}

func isHiddenName(name string) bool {
	return strings.HasPrefix(name, ".")
}
