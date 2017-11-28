package netconsoled

import (
	"fmt"
)

// A Sink enables storage of processed logs.
type Sink interface {
	// Store stores a log to the Sink.
	Store(d Data) error

	// String returns the name of a Sink.
	fmt.Stringer
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
