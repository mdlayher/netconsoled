package netconsoled

import (
	"fmt"
	"net"

	"github.com/mdlayher/netconsole"
)

// A Filter allows filtering of incoming logs based on the source network address
// and contents of the logs.
type Filter interface {
	// Allow determines if a log should or should not be processed.
	Allow(addr net.Addr, l *netconsole.Log) bool

	// String returns the name of a Filter.
	fmt.Stringer
}

// MultiFilter chains zero or more Filters together.  If any Filter does not allow
// a given log, subsequent Filters in the chain are not invoked.
func MultiFilter(filters ...Filter) Filter {
	return &multiFilter{
		filters: filters,
	}
}

var _ Filter = &multiFilter{}

type multiFilter struct {
	filters []Filter
}

func (f *multiFilter) Allow(addr net.Addr, l *netconsole.Log) bool {
	for _, filter := range f.filters {
		if !filter.Allow(addr, l) {
			return false
		}
	}

	return true
}

func (f *multiFilter) String() string {
	// TODO(mdlayher): loop through filters and list.
	return "multi"
}

// FuncFilter adapts a function into a Filter.
func FuncFilter(allow func(addr net.Addr, l *netconsole.Log) bool) Filter {
	return &funcFilter{
		fn: allow,
	}
}

var _ Filter = &funcFilter{}

type funcFilter struct {
	fn func(_ net.Addr, _ *netconsole.Log) bool
}

func (f *funcFilter) Allow(addr net.Addr, l *netconsole.Log) bool { return f.fn(addr, l) }
func (f *funcFilter) String() string                              { return "func" }

// NoopFilter returns a Filter that always allows any log.
func NoopFilter() Filter {
	return &noopFilter{}
}

var _ Filter = &noopFilter{}

type noopFilter struct{}

func (f *noopFilter) Allow(_ net.Addr, _ *netconsole.Log) bool { return true }
func (f *noopFilter) String() string                           { return "noop" }
