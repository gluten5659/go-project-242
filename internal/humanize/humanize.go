package humanize

import "fmt"

var sizes = []string{
	"B",
	"KB",
	"MB",
	"GB",
	"TB",
	"PB",
	"EB",
}

func Format(byteCount int64, formatNeeded bool) string {
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

func Line(output, path string) string {
	return fmt.Sprintf("%s\t%s", output, path)
}
