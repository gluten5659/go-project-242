package code

import (
	"code/internal/humanize"
	"code/internal/scan"
)

func GetPathSize(path string, recursive bool, formatNeeded bool, listHidden bool) (string, error) {
	size, err := scan.Measure(path, listHidden, recursive)
	if err != nil {
		return "", err
	}
	return humanize.Format(size, formatNeeded), nil
}
