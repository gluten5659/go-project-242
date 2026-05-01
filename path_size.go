package code

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	ListHidden   bool
	Recursive    bool
	FormatNeeded bool
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

func FormatedSize(path string) string {
	size := float64(GetPathSize(path))
	prefix := 0
	if !FormatNeeded {
		return fmt.Sprintf("%.0fB	%s", size, path)
	}
	for size > 1023.9 {
		prefix++
		size = size / 1024
	}
	return fmt.Sprintf("%.1f%s	%s", size, Sizes[prefix], path)
}

func GetPathSize(path string) int {
	stat, _ := os.Lstat(path)
	size := 0
	if stat.IsDir() {
		size += getFolderSize(path)
	} else {
		size = int(stat.Size())
	}
	return size
}

func getFolderSize(folderPath string) int {
	files, _ := os.ReadDir(folderPath)
	folderSize := 0
	for _, file := range files {
		if !ListHidden && file.Name()[0] == '.' {
			continue
		}
		if !Recursive && file.IsDir() {
			continue
		}
		folderSize += GetPathSize(filepath.Join(folderPath, file.Name()))
	}
	return folderSize
}
