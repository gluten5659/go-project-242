package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

var listHidden bool

func main() {
	var formatNeeded bool
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "human",
				Usage:       "Converts B into more readable KB/MB/GB etc.",
				Aliases:     []string{"H"},
				Destination: &formatNeeded,
			},
			&cli.BoolFlag{
				Name:        "all",
				Usage:       "Allow hidden files",
				Aliases:     []string{"a"},
				Destination: &listHidden,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			path := cmd.Args().Get(0)
			fmt.Println(FormatedSize(path, formatNeeded))

			return nil
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

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

func FormatedSize(path string, formatNeeded bool) string {
	size := float64(GetSize(path))
	prefix := 0
	if !formatNeeded {
		return fmt.Sprintf("%.0fB	%s", size, path)
	}
	for size > 1023.9 {
		prefix++
		size = size / 1024
	}
	return fmt.Sprintf("%.1f%s	%s", size, Sizes[prefix], path)
}

func GetSize(path string) int {
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
		if !listHidden && file.Name()[0] == '.' {
			continue
		}
		folderSize += GetSize(folderPath + `/` + file.Name())
	}
	return folderSize
}
