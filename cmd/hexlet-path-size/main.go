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
			fmt.Printf("%dB	%s\n", GetSize(path), path)

			return nil
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
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
		folderSize += GetSize(folderPath + `/` + file.Name())
	}
	return folderSize
}
