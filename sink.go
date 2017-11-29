package netconsoled

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// defaultFormat is the default format descriptor for Sinks.
const defaultFormat = "[% 15s] [% 15f] %s"

// A Sink enables storage of processed logs.
//
// Sinks may optionally implement io.Closer to flush data before the server halts.
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

func (s *multiSink) Close() error {
	for _, sink := range s.sinks {
		// Close all sinks which implement io.Closer.
		c, ok := sink.(io.Closer)
		if !ok {
			continue
		}

		if err := c.Close(); err != nil {
			return err
		}
	}

	return nil
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

// FileSink creates a Sink that creates or opens the specified file and appends
// logs to the file.
func FileSink(file string) (Sink, error) {
	file = filepath.Clean(file)

	// Create or open the file, and always append to it.
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return newNamedSink(fmt.Sprintf("file: %q", file), WriterSink(f)), nil
}

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

// A syncer is a type which can flush its contents from memory to disk, e.g.
// an *os.File.
type syncer interface {
	Sync() error
}

var _ syncer = &os.File{}

func (s *writerSink) Close() error {
	// Since a writerSink can be used for files, it's possible that the io.Writer
	// has a Sync method to flush its contents to disk.  Try it first.
	if sync, ok := s.w.(syncer); ok {
		// Attempting to sync stdout, at least on Linux, results in
		// "invalid argument".  Instead of doing build tags and OS-specific
		// checks, keep it simple and just Sync as a best effort.
		_ = sync.Sync()
	}

	// Close io.Writers which also implement io.Closer.
	c, ok := s.w.(io.Closer)
	if !ok {
		return nil
	}

	return c.Close()
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

func (s *namedSink) Close() error {
	// Close Sinks which also implement io.Closer.
	c, ok := s.sink.(io.Closer)
	if !ok {
		return nil
	}

	return c.Close()
}
func (s *namedSink) Store(d Data) error { return s.sink.Store(d) }
func (s *namedSink) String() string     { return s.name }
