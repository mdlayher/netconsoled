package netconsoled_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netconsole"
	"github.com/mdlayher/netconsoled"
)

var panicFilter = netconsoled.FuncFilter(func(d netconsoled.Data) bool {
	panic("reached panic filter")
})

func TestFilter(t *testing.T) {
	tests := []struct {
		name   string
		d      netconsoled.Data
		verify func(t *testing.T, d netconsoled.Data)
	}{
		{
			name:   "multi disallow",
			verify: testMultiFilterDisallow,
		},
		{
			name: "multi allow",
			d: netconsoled.Data{
				Log: netconsole.Log{
					Elapsed: 1 * time.Second,
					Message: "hello world",
				},
			},
			verify: testMultiFilterAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.verify(t, tt.d)
		})
	}
}

func testMultiFilterDisallow(t *testing.T, d netconsoled.Data) {
	t.Helper()

	disallowFilter := netconsoled.FuncFilter(func(d netconsoled.Data) bool {
		return false
	})

	filter := netconsoled.MultiFilter(
		netconsoled.NoopFilter(),
		// Should stop here and not reach panic filter.
		disallowFilter,
		panicFilter,
	)

	if filter.Allow(d) {
		t.Fatal("expected filter to disallow log, but it was allowed")
	}
}

func testMultiFilterAllow(t *testing.T, d netconsoled.Data) {
	t.Helper()

	var got []netconsoled.Data
	fnFilter := netconsoled.FuncFilter(func(d netconsoled.Data) bool {
		got = append(got, d)
		return true
	})

	filter := netconsoled.MultiFilter(
		netconsoled.NoopFilter(),
		fnFilter,
		fnFilter,
	)

	if !filter.Allow(d) {
		t.Fatal("expected filter to allow log, but it was disallowed")
	}

	want := []netconsoled.Data{d, d}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected logs (-want +got):\n%s", diff)
	}
}
