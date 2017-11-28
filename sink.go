package netconsoled

import (
	"fmt"
	"io"
	"os"
)

// defaultFormat is the default format descriptor for Sinks.
const defaultFormat = "[% 15s] [% 15f] %s"

// A Sink enables storage of processed logs.
type Sink interface {
	// Store stores a log to the Sink.
	Store(d Data) error

	// String returns the name of a Sink.
	fmt.Stringer
}

// StdoutSink creates a Sink that writes log data to stdout.
func StdoutSink() Sink {
	return newNamedSink("stdout", WriterSink(os.Stdout))
}

// MultiSink chains zero or more Sinks together.  If any Sink returns an error,
// subsequent Sinks in the chain are not invoked.
func MultiSink(sinks ...Sink) Sink {
	return &multiSink{
		sinks: sinks,
	}
}

var _ Sink = &multiSink{}

type multiSink struct {
	sinks []Sink
}

func (s *multiSink) Store(d Data) error {
	for _, sink := range s.sinks {
		if err := sink.Store(d); err != nil {
			return err
		}
	}

	return nil
}

func (s *multiSink) String() string {
	// TODO(mdlayher): loop through sinks and list.
	return "multi"
}

// FuncSink adapts a function into a Sink.
func FuncSink(store func(d Data) error) Sink {
	return &funcSink{
		fn: store,
	}
}

var _ Sink = &funcSink{}

type funcSink struct {
	fn func(d Data) error
}

func (f *funcSink) Store(d Data) error { return f.fn(d) }
func (f *funcSink) String() string     { return "func" }

// NoopSink returns a Sink that discards all logs.
func NoopSink() Sink {
	return &noopSink{}
}

var _ Sink = &noopSink{}

type noopSink struct{}

func (s *noopSink) Store(_ Data) error { return nil }
func (s *noopSink) String() string     { return "noop" }

// WriterSink creates a Sink that writes to w.
func WriterSink(w io.Writer) Sink {
	return &writerSink{
		w: w,
		// Add a newline on behalf of the caller for ease of use.
		// TODO(mdlayher): expose formatting later?
		format: defaultFormat + "\n",
	}
}

var _ Sink = &writerSink{}

type writerSink struct {
	w      io.Writer
	format string
}

func (s *writerSink) Store(d Data) error {
	_, err := fmt.Fprintf(s.w, s.format, d.Addr, d.Log.Elapsed.Seconds(), d.Log.Message)
	return err
}

func (s *writerSink) String() string { return "writer" }

// newNamedSink wraps a Sink and replaces its name with the specified name.
// This is primarily useful for composing Sinks and providing detailed information
// to the user on startup.
func newNamedSink(name string, sink Sink) Sink {
	return &namedSink{
		sink: sink,
		name: name,
	}
}

var _ Sink = &namedSink{}

type namedSink struct {
	sink Sink
	name string
}

func (s *namedSink) Store(d Data) error { return s.sink.Store(d) }
func (s *namedSink) String() string     { return s.name }
