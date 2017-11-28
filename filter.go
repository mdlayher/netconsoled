package netconsoled

import (
	"fmt"
)

// A Filter allows filtering and transformation of incoming logs based on the
// contents of the logs.
type Filter interface {
	// Filter determines if Data should proceed to the next step in a pipeline,
	// and may also apply transformations to the Data.
	//
	// If Data should not continue in the pipeline, pass is false.  Any error
	// will also result in Data not being passed down the pipeline.
	Filter(in Data) (out Data, pass bool, err error)

	// String returns the name of a Filter.
	fmt.Stringer
}

// MultiFilter chains zero or more Filters together.  The output of each Filter
// is passed to the next Filter in the chain.  If any Filter does not pass
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

func (f *multiFilter) Filter(in Data) (Data, bool, error) {
	var (
		out  Data
		pass bool
		err  error
	)

	for _, filter := range f.filters {
		out, pass, err = filter.Filter(in)
		if err != nil {
			return Data{}, false, err
		}
		if !pass {
			return Data{}, false, nil
		}

		in = out
	}

	return out, true, nil
}

func (f *multiFilter) String() string {
	// TODO(mdlayher): loop through filters and list.
	return "multi"
}

// FuncFilter adapts a function into a Filter.
func FuncFilter(filter func(in Data) (Data, bool, error)) Filter {
	return &funcFilter{
		fn: filter,
	}
}

var _ Filter = &funcFilter{}

type funcFilter struct {
	fn func(in Data) (Data, bool, error)
}

func (f *funcFilter) Filter(in Data) (Data, bool, error) { return f.fn(in) }
func (f *funcFilter) String() string                     { return "func" }

// NoopFilter returns a Filter that performs no processing and always passes a log.
func NoopFilter() Filter {
	return &noopFilter{}
}

var _ Filter = &noopFilter{}

type noopFilter struct{}

func (f *noopFilter) Filter(in Data) (Data, bool, error) { return in, true, nil }
func (f *noopFilter) String() string                     { return "noop" }
