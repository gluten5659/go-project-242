package code

import (
	"fmt"
	"os"
	"path/filepath"
)

var Sizes = []string{
	"B",
	"KB",
	"MB",
	"GB",
	"TB",
	"PB",
	"EB",
	"PLZ NO MORE",
}

func GetPathSize(path string, recursive bool, listHidden bool, formatNeeded bool) (string, error) {
	size, err := getSize(path, listHidden, recursive)
	if err != nil {
		return "", err
	}

	fsize := float64(size)
	prefix := 0
	if !formatNeeded {
		return fmt.Sprintf("%.0fB", fsize), nil
	}
	for fsize > 1023.9 {
		prefix++
		fsize = fsize / 1024
	}
	return fmt.Sprintf("%.1f%s", fsize, Sizes[prefix]), nil
}

func getSize(path string, listHidden bool, recursive bool) (int, error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	if stat.IsDir() {
		return getFolderSize(path, listHidden, recursive)
	}
	return int(stat.Size()), nil
}

func getFolderSize(folderPath string, listHidden bool, recursive bool) (int, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return 0, err
	}
	folderSize := 0
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
