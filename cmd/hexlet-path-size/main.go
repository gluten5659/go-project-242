package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Action: func(ctx context.Context, cmd *cli.Command) error {
			path := cmd.Args().Get(0)
			fmt.Printf("%dB	%s", GetSize(path), path)

			return nil
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func GetSize(path string) int {
	size, err := getFileSize(path)
	if err != nil {
		size, _ = getFolderSize(path)
	}
	return size
}

func getFileSize(filePath string) (int, error) {
	stat, err := os.Lstat(filePath)
	if err != nil {
		return 0, err
	}
	return int(stat.Size()), nil
}

func getFolderSize(folderPath string) (int, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return 0, err
	}
	folderSize := 0
	for _, file := range files {
		if file.Type().IsDir() {
			continue
		}
		fileInfo, err := file.Info()
		if err != nil {
			return 0, err
		}
		folderSize += int(fileInfo.Size())
	}
	return folderSize, nil
}
