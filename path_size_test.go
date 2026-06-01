package code

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"code/internal/testutil"
)

func TestGetPathSize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc          string
		setup         func(t *testing.T) string
		recursive     bool
		formatNeeded  bool
		includeHidden bool
		want          string
		wantErr       bool
		wantErrIs     error
	}{
		{
			desc:         "raw bytes for file",
			setup:        testutil.TempFile("a.txt", "hello"),
			formatNeeded: false,
			want:         "5B",
		},
		{
			desc:         "formatted KB for file",
			setup:        testutil.TempFile("big.dat", strings.Repeat("\x00", 1024*3/2)),
			formatNeeded: true,
			want:         "1.5KB",
		},
		{
			desc:         "recursive directory total",
			setup:        testutil.NestedTree("hello", "world!"),
			recursive:    true,
			formatNeeded: false,
			want:         "11B",
		},
		{
			desc:      "nonexistent path returns error",
			setup:     testutil.StaticPath("/no/such/path"),
			wantErr:   true,
			wantErrIs: fs.ErrNotExist,
		},
		{
			desc:          "hidden file path is shown despite includeHidden being false",
			setup:         testutil.TempFile(".env", "PORT=8080"),
			formatNeeded:  false,
			includeHidden: false,
			want:          "9B",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()
			path := tC.setup(t)

			got, err := GetPathSize(path, tC.recursive, tC.formatNeeded, tC.includeHidden)
			if (err != nil) != tC.wantErr {
				t.Fatalf("GetPathSize error = %v, wantErr %v", err, tC.wantErr)
			}

			if tC.wantErrIs != nil && !errors.Is(err, tC.wantErrIs) {
				t.Errorf("GetPathSize error = %v, want errors.Is(_, %v)", err, tC.wantErrIs)
			}

			if got != tC.want {
				t.Errorf("GetPathSize = %q, want %q", got, tC.want)
			}
		})
	}
}
