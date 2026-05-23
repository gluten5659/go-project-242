package humanize

import "fmt"

const unitBase = 1024

var unitSuffixes = []string{
	"B",
	"KB",
	"MB",
	"GB",
	"TB",
	"PB",
	"EB",
}

func Format(byteCount int64, formatNeeded bool) string {
	if !formatNeeded || byteCount < unitBase {
		return fmt.Sprintf("%dB", byteCount)
	}

	value, suffix := scaleToUnit(byteCount)

	return fmt.Sprintf("%.1f%s", value, suffix)
}

func scaleToUnit(byteCount int64) (float64, string) {
	value := float64(byteCount)
	for _, suffix := range unitSuffixes {
		if value < unitBase {
			return value, suffix
		}

		value /= unitBase
	}

	return value, unitSuffixes[len(unitSuffixes)-1]
}

func FormatLine(output, path string) string {
	return fmt.Sprintf("%s\t%s", output, path)
}
