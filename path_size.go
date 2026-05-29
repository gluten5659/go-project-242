package code

import (
	"code/internal/dirsize"
	"code/internal/output"
)

func GetPathSize(path string, recursive bool, formatNeeded bool, listHidden bool) (string, error) {
	size, err := dirsize.Measure(path, listHidden, recursive)
	if err != nil {
		return "", err
	}

	return output.FormatSize(size, formatNeeded), nil
}
