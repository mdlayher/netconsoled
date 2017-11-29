package netconsoled_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
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
		{
			name: "writer ok",
			d: netconsoled.Data{
				Addr: &net.UDPAddr{
					IP:   net.IPv4(192, 168, 1, 1),
					Port: 6666,
				},
				Log: netconsole.Log{
					Message: "hello world",
				},
			},
			verify: testWriterSinkOK,
		},
		{
			name:   "multi closer ok",
			verify: testMultiSinkCloserOK,
		},
		{
			name: "file ok",
			d: netconsoled.Data{
				Log: netconsole.Log{
					Message: "hello world",
				},
			},
			verify: testFileSinkOK,
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

func testWriterSinkOK(t *testing.T, d netconsoled.Data) {
	t.Helper()

	buf := bytes.NewBuffer(nil)
	sink := netconsoled.WriterSink(buf)

	if err := sink.Store(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Just check if a couple of string fields ended up in the buffer.
	ss := []string{
		d.Addr.String(),
		"0.0",
		d.Log.Message,
	}

	str := buf.String()
	for _, s := range ss {
		if !strings.Contains(str, s) {
			t.Fatalf("buffer did not contain %q: buf: %s", s, str)
		}
	}
}

func testMultiSinkCloserOK(t *testing.T, d netconsoled.Data) {
	t.Helper()

	// Attach a Close method to verify an arbitrary Sink is closed,
	// though the underlying type doesn't have a Close method.
	fnSink := &sinkCloser{
		Sink: netconsoled.FuncSink(func(d netconsoled.Data) error {
			return nil
		}),
	}

	// Do not attach a Close method to a writer sink.
	buf := bytes.NewBuffer(nil)
	wSink := netconsoled.WriterSink(buf)

	sink := netconsoled.MultiSink(
		fnSink,
		wSink,
	)

	if err := sink.Store(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, ok := sink.(io.Closer)
	if !ok {
		t.Fatal("multi sink is not an io.Closer")
	}

	if err := c.Close(); err != nil {
		t.Fatalf("failed to close sink: %v", err)
	}

	if !fnSink.closed {
		t.Fatal("function sink Close was not called")
	}
}

type sinkCloser struct {
	netconsoled.Sink
	closed bool
}

func (s *sinkCloser) Close() error {
	s.closed = true
	return nil
}

func testFileSinkOK(t *testing.T, d netconsoled.Data) {
	// Ensure a test file doesn't already exist, but also clean it
	// up after the test.
	file := filepath.Join(os.TempDir(), "netconsoled_filesink.tmp")
	_ = os.Remove(file)
	defer os.Remove(file)

	// Open, write data, and close the sink twice.
	// Verify that the same log was written twice, e.g. the file was not
	// truncated after the second open.
	for i := 0; i < 2; i++ {
		sink, err := netconsoled.FileSink(file)
		if err != nil {
			t.Fatalf("failed to create test file sink: %v", err)
		}

		if err := sink.Store(d); err != nil {
			t.Fatalf("failed to perform write %d: %v", i, err)
		}

		// If this panics, we've got a bigger problem anyway.
		_ = sink.(io.Closer).Close()
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if c := bytes.Count(b, []byte(d.Log.Message)); c != 2 {
		t.Fatalf("log was not written %d times to buffer, not 2 times: %s", c, string(b))
	}
}
