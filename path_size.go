package code

import (
	"code/internal/dirsize"
	"code/internal/output"
)

func GetPathSize(path string, recursive bool, formatNeeded bool, includeHidden bool) (string, error) {
	size, err := dirsize.Measure(path, includeHidden, recursive)
	if err != nil {
		return "", err
	}

	return output.FormatSize(size, formatNeeded), nil
}
