package netconsoled_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netconsole"
	"github.com/mdlayher/netconsoled"
)

var panicFilter = netconsoled.FuncFilter(func(in netconsoled.Data) (netconsoled.Data, bool, error) {
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

	disallowFilter := netconsoled.FuncFilter(func(in netconsoled.Data) (netconsoled.Data, bool, error) {
		return netconsoled.Data{}, false, nil
	})

	filter := netconsoled.MultiFilter(
		netconsoled.NoopFilter(),
		// Should stop here and not reach panic filter.
		disallowFilter,
		panicFilter,
	)

	_, pass, err := filter.Filter(d)
	if err != nil {
		t.Fatalf("failed to filter log: %v", err)
	}
	if pass {
		t.Fatal("expected filter to disallow log, but it was allowed")
	}
}

func testMultiFilterAllow(t *testing.T, d netconsoled.Data) {
	t.Helper()

	var got []netconsoled.Data
	fnFilter := netconsoled.FuncFilter(func(in netconsoled.Data) (netconsoled.Data, bool, error) {
		got = append(got, d)
		return in, true, nil
	})

	filter := netconsoled.MultiFilter(
		netconsoled.NoopFilter(),
		fnFilter,
		fnFilter,
	)

	out, pass, err := filter.Filter(d)
	if err != nil {
		t.Fatalf("failed to filter log: %v", err)
	}
	if !pass {
		t.Fatal("expected filter to allow log, but it was disallowed")
	}

	if diff := cmp.Diff(out, d); diff != "" {
		t.Fatalf("unexpected filtered data (-want +got):\n%s", diff)
	}

	want := []netconsoled.Data{d, d}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected logs (-want +got):\n%s", diff)
	}
}
