package netconsoled_test

import (
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netconsole"
	"github.com/mdlayher/netconsoled"
)

var panicFilter = netconsoled.FuncFilter(func(addr net.Addr, l *netconsole.Log) bool {
	panic("reached panic filter")
})

func TestFilter(t *testing.T) {
	tests := []struct {
		name   string
		addr   net.Addr
		l      *netconsole.Log
		verify func(t *testing.T, addr net.Addr, l *netconsole.Log)
	}{
		{
			name:   "multi disallow",
			verify: testMultiFilterDisallow,
		},
		{
			name: "multi allow",
			l: &netconsole.Log{
				Elapsed: 1 * time.Second,
				Message: "hello world",
			},
			verify: testMultiFilterAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.verify(t, tt.addr, tt.l)
		})
	}
}

func testMultiFilterDisallow(t *testing.T, addr net.Addr, l *netconsole.Log) {
	t.Helper()

	disallowFilter := netconsoled.FuncFilter(func(addr net.Addr, l *netconsole.Log) bool {
		return false
	})

	filter := netconsoled.MultiFilter(
		netconsoled.NoopFilter(),
		// Should stop here and not reach panic filter.
		disallowFilter,
		panicFilter,
	)

	if filter.Allow(addr, l) {
		t.Fatal("expected filter to disallow log, but it was allowed")
	}
}

func testMultiFilterAllow(t *testing.T, addr net.Addr, l *netconsole.Log) {
	t.Helper()

	var got []*netconsole.Log
	fnFilter := netconsoled.FuncFilter(func(addr net.Addr, l *netconsole.Log) bool {
		got = append(got, l)
		return true
	})

	filter := netconsoled.MultiFilter(
		netconsoled.NoopFilter(),
		fnFilter,
		fnFilter,
	)

	if !filter.Allow(addr, l) {
		t.Fatal("expected filter to allow log, but it was disallowed")
	}

	want := []*netconsole.Log{l, l}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected logs (-want +got):\n%s", diff)
	}
}
