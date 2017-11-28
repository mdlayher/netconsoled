package netconsoled_test

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netconsole"
	"github.com/mdlayher/netconsoled"
)

var panicSink = netconsoled.FuncSink(func(addr net.Addr, l *netconsole.Log) error {
	panic("reached panic sink")
})

func TestSink(t *testing.T) {
	tests := []struct {
		name   string
		addr   net.Addr
		l      *netconsole.Log
		verify func(t *testing.T, addr net.Addr, l *netconsole.Log)
	}{
		{
			name:   "multi error",
			verify: testMultiSinkError,
		},

		{
			name: "multi ok",
			l: &netconsole.Log{
				Elapsed: 1 * time.Second,
				Message: "hello world",
			},
			verify: testMultiSinkOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.verify(t, tt.addr, tt.l)
		})
	}
}

func testMultiSinkError(t *testing.T, addr net.Addr, l *netconsole.Log) {
	t.Helper()

	errSink := netconsoled.FuncSink(func(addr net.Addr, l *netconsole.Log) error {
		return errors.New("some error")
	})

	sink := netconsoled.MultiSink(
		netconsoled.NoopSink(),
		// Should stop here and not reach panic sink.
		errSink,
		panicSink,
	)

	if err := sink.Store(addr, l); err == nil {
		t.Fatal("expected an error, but none occurred")
	}
}

func testMultiSinkOK(t *testing.T, addr net.Addr, l *netconsole.Log) {
	t.Helper()

	var got []*netconsole.Log
	fnSink := netconsoled.FuncSink(func(addr net.Addr, l *netconsole.Log) error {
		got = append(got, l)
		return nil
	})

	sink := netconsoled.MultiSink(
		netconsoled.NoopSink(),
		fnSink,
		fnSink,
	)

	if err := sink.Store(addr, l); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []*netconsole.Log{l, l}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected logs (-want +got):\n%s", diff)
	}
}
