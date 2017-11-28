package netconsoled

import (
	"fmt"
	"net"

	"github.com/mdlayher/netconsole"
)

// A Sink enables storage of processed logs.
type Sink interface {
	// Store stores a log to the Sink.
	Store(addr net.Addr, l *netconsole.Log) error

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

func (s *multiSink) Store(addr net.Addr, l *netconsole.Log) error {
	for _, sink := range s.sinks {
		if err := sink.Store(addr, l); err != nil {
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
func FuncSink(store func(addr net.Addr, l *netconsole.Log) error) Sink {
	return &funcSink{
		fn: store,
	}
}

var _ Sink = &funcSink{}

type funcSink struct {
	fn func(_ net.Addr, _ *netconsole.Log) error
}

func (f *funcSink) Store(addr net.Addr, l *netconsole.Log) error { return f.fn(addr, l) }
func (f *funcSink) String() string                               { return "func" }

// NoopSink returns a Sink that discards all logs.
func NoopSink() Sink {
	return &noopSink{}
}

var _ Sink = &noopSink{}

type noopSink struct{}

func (s *noopSink) Store(_ net.Addr, _ *netconsole.Log) error { return nil }
func (s *noopSink) String() string                            { return "noop" }
