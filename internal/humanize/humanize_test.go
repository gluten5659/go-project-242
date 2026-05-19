package humanize

import "testing"

func TestFormat(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc         string
		byteCount    int64
		formatNeeded bool
		want         string
	}{
		{"raw bytes when format disabled", 1500, false, "1500B"},
		{"under 1KB stays as bytes", 500, true, "500B"},
		{"exactly 1023 bytes stays as bytes", 1024 - 1, true, "1023B"},
		{"exactly 1KB", 1024, true, "1.0KB"},
		{"1.5 MB", 1024 * 1024 * 3 / 2, true, "1.5MB"},
		{"2 GB", 1024 * 1024 * 1024 * 2, true, "2.0GB"},
		{"zero", 0, true, "0B"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()
			got := Format(tC.byteCount, tC.formatNeeded)
			if got != tC.want {
				t.Errorf("Format(%d, %v) = %q, want %q",
					tC.byteCount, tC.formatNeeded, got, tC.want)
			}
		})
	}
}

func TestPickUnit(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc      string
		byteCount int64
		wantValue float64
		wantUnit  string
	}{
		{"zero", 0, 0, "B"},
		{"1023", 1024 - 1, 1023, "B"},
		{"1024", 1024, 1, "KB"},
		{"1.5 KB", 1024 * 3 / 2, 1.5, "KB"},
		{"1 EB", 1024 * 1024 * 1024 * 1024 * 1024 * 1024, 1, "EB"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()
			gotValue, gotUnit := pickUnit(tC.byteCount)
			if gotValue != tC.wantValue || gotUnit != tC.wantUnit {
				t.Errorf("pickUnit(%d) = (%v, %q), want (%v, %q)",
					tC.byteCount, gotValue, gotUnit,
					tC.wantValue, tC.wantUnit)
			}
		})
	}
}

func TestLine(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc   string
		output string
		path   string
		want   string
	}{
		{"size and path joined by tab", "5B", "/tmp/a.txt", "5B\t/tmp/a.txt"},
		{"human-readable size", "1.5KB", "/var/big.dat", "1.5KB\t/var/big.dat"},
		{"empty path", "0B", "", "0B\t"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()
			got := Line(tC.output, tC.path)
			if got != tC.want {
				t.Errorf("Line(%q, %q) = %q, want %q",
					tC.output, tC.path, got, tC.want)
			}
		})
	}
}
