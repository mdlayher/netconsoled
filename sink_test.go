package netconsoled_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netconsole"
	"github.com/mdlayher/netconsoled"
)

var panicSink = netconsoled.FuncSink(func(d netconsoled.Data) error {
	panic("reached panic sink")
})

func TestSink(t *testing.T) {
	tests := []struct {
		name   string
		d      netconsoled.Data
		verify func(t *testing.T, d netconsoled.Data)
	}{
		{
			name:   "multi error",
			verify: testMultiSinkError,
		},

		{
			name: "multi ok",
			d: netconsoled.Data{
				Log: netconsole.Log{
					Elapsed: 1 * time.Second,
					Message: "hello world",
				},
			},
			verify: testMultiSinkOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.verify(t, tt.d)
		})
	}
}

func testMultiSinkError(t *testing.T, d netconsoled.Data) {
	t.Helper()

	errSink := netconsoled.FuncSink(func(d netconsoled.Data) error {
		return errors.New("some error")
	})

	sink := netconsoled.MultiSink(
		netconsoled.NoopSink(),
		// Should stop here and not reach panic sink.
		errSink,
		panicSink,
	)

	if err := sink.Store(d); err == nil {
		t.Fatal("expected an error, but none occurred")
	}
}

func testMultiSinkOK(t *testing.T, d netconsoled.Data) {
	t.Helper()

	var got []netconsoled.Data
	fnSink := netconsoled.FuncSink(func(d netconsoled.Data) error {
		got = append(got, d)
		return nil
	})

	sink := netconsoled.MultiSink(
		netconsoled.NoopSink(),
		fnSink,
		fnSink,
	)

	if err := sink.Store(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []netconsoled.Data{d, d}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected logs (-want +got):\n%s", diff)
	}
}
