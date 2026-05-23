package scan

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathNotFound     = errors.New("path not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrUnsupportedPath  = errors.New("unsupported path type")
	ErrReadFailed       = errors.New("read failed")
)

func wrapFSError(err error, path string) error {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("%w: %q: %w", ErrPathNotFound, path, err)
	case errors.Is(err, fs.ErrPermission):
		return fmt.Errorf("%w: %q: %w", ErrPermissionDenied, path, err)
	default:
		return fmt.Errorf("%w: %q: %w", ErrReadFailed, path, err)
	}
}

func Measure(path string, includeHidden, recursive bool) (int64, error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return 0, wrapFSError(err, path)
	}

	mode := stat.Mode()
	switch {
	case mode.IsDir():
		return measureFolder(path, includeHidden, recursive)
	case mode.IsRegular():
		return stat.Size(), nil
	case mode&os.ModeSymlink != 0:
		return stat.Size(), nil
	default:
		return 0, fmt.Errorf("%w: %q", ErrUnsupportedPath, path)
	}
}

func measureFolder(folderPath string, includeHidden bool, recursive bool) (int64, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return 0, wrapFSError(err, folderPath)
	}

	var folderSize int64

	for _, file := range files {
		if !includeHidden && isHiddenName(file.Name()) {
			continue
		}

		if !recursive && file.IsDir() {
			continue
		}

		childPath := filepath.Join(folderPath, file.Name())

		size, err := Measure(childPath, includeHidden, recursive)
		if err != nil {
			return 0, err
		}

		folderSize += size
	}

	return folderSize, nil
}

func isHiddenName(name string) bool {
	return strings.HasPrefix(name, ".")
}
