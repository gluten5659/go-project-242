package code

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var (
	ErrPathNotFound     = errors.New("path not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrUnsupportedPath  = errors.New("unsupported path type")
	ErrReadFailed       = errors.New("read failed")
)

var sizes = []string{
	"B",
	"KB",
	"MB",
	"GB",
	"TB",
	"PB",
	"EB",
}

func wrapFSError(err error, path string) error {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("%w: %q", ErrPathNotFound, path)
	case errors.Is(err, fs.ErrPermission):
		return fmt.Errorf("%w: %q", ErrPermissionDenied, path)
	default:
		return fmt.Errorf("%w: %q: %w", ErrReadFailed, path, err)
	}
}

func GetPathSize(path string, recursive bool, formatNeeded bool, listHidden bool) (string, error) {
	size, err := getSize(path, listHidden, recursive)
	if err != nil {
		return "", err
	}
	return formatOutput(size, formatNeeded), nil
}

func formatOutput(byteCount int64, formatNeeded bool) string {
	floatSize, prefix := pickUnit(byteCount)
	if !formatNeeded || prefix == "B" {
		return fmt.Sprintf("%dB", byteCount)
	}
	return fmt.Sprintf("%.1f%s", floatSize, prefix)
}

func pickUnit(byteCount int64) (float64, string) {
	floatBytesCount := float64(byteCount)
	prefixIndex := 0
	for floatBytesCount >= 1024 {
		prefixIndex++
		floatBytesCount /= 1024
	}
	return floatBytesCount, sizes[prefixIndex]
}

func getSize(path string, listHidden bool, recursive bool) (int64, error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return 0, wrapFSError(err, path)
	}
	mode := stat.Mode()
	switch {
	case mode.IsDir():
		return getFolderSize(path, listHidden, recursive)
	case mode.IsRegular(), mode&os.ModeSymlink != 0:
		return stat.Size(), nil
	default:
		return 0, fmt.Errorf("%w: %q", ErrUnsupportedPath, path)
	}
}

func getFolderSize(folderPath string, listHidden bool, recursive bool) (int64, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return 0, wrapFSError(err, folderPath)
	}
	var folderSize int64
	for _, file := range files {
		if !listHidden && file.Name()[0] == '.' {
			continue
		}
		if !recursive && file.IsDir() {
			continue
		}
		size, err := getSize(filepath.Join(folderPath, file.Name()), listHidden, recursive)
		if err != nil {
			return 0, err
		}
		folderSize += size
	}
	return folderSize, nil
}
