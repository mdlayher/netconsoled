package netconsoled

import (
	"fmt"
)

// A Filter allows filtering of incoming logs based on the source network address
// and contents of the logs.
type Filter interface {
	// Allow determines if Data should or should not be processed.
	Allow(d Data) bool

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

func (f *multiFilter) Allow(d Data) bool {
	for _, filter := range f.filters {
		if !filter.Allow(d) {
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
func FuncFilter(allow func(d Data) bool) Filter {
	return &funcFilter{
		fn: allow,
	}
}

var _ Filter = &funcFilter{}

type funcFilter struct {
	fn func(d Data) bool
}

func (f *funcFilter) Allow(d Data) bool { return f.fn(d) }
func (f *funcFilter) String() string    { return "func" }

// NoopFilter returns a Filter that always allows any log.
func NoopFilter() Filter {
	return &noopFilter{}
}

var _ Filter = &noopFilter{}

type noopFilter struct{}

func (f *noopFilter) Allow(_ Data) bool { return true }
func (f *noopFilter) String() string    { return "noop" }
