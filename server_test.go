package netconsoled_test

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netconsole"
	"github.com/mdlayher/netconsoled"
)

func TestServerHandle(t *testing.T) {
	tests := []struct {
		name   string
		addr   net.Addr
		l      netconsole.Log
		verify func(t *testing.T, addr net.Addr, l netconsole.Log)
	}{
		{
			name: "filter disallow",
			addr: &net.UDPAddr{
				IP:   net.IPv4(192, 168, 1, 1),
				Port: 6666,
			},
			verify: testServerFilterDisallow,
		},
		{
			name: "sink ok",
			addr: &net.UDPAddr{
				IP:   net.IPv4(192, 168, 1, 1),
				Port: 6666,
			},
			verify: testServerSinkOK,
		},
		{
			name: "metrics ok",
			addr: &net.UDPAddr{
				IP:   net.IPv4(192, 168, 1, 1),
				Port: 6666,
			},
			verify: testServerMetricsOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.verify(t, tt.addr, tt.l)
		})
	}
}

func testServerFilterDisallow(t *testing.T, addr net.Addr, l netconsole.Log) {
	t.Helper()

	s := &netconsoled.Server{
		Filter: netconsoled.FuncFilter(func(d netconsoled.Data) (netconsoled.Data, bool, error) {
			return netconsoled.Data{}, false, nil
		}),
		Sink: panicSink,
	}

	s.Handle(addr, l)
}

func testServerSinkOK(t *testing.T, addr net.Addr, l netconsole.Log) {
	t.Helper()

	s := &netconsoled.Server{
		Filter: netconsoled.NoopFilter(),
		Sink:   netconsoled.NoopSink(),
	}

	s.Handle(addr, l)
}

func testServerMetricsOK(t *testing.T, addr net.Addr, l netconsole.Log) {
	t.Helper()

	metrics, reg := netconsoled.NewMetrics()

	s := &netconsoled.Server{
		Filter:  netconsoled.NoopFilter(),
		Sink:    netconsoled.NoopSink(),
		Metrics: metrics,
	}

	s.Handle(addr, l)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Assume that each phase of the processing pipeline has been hit
	// exactly once.
	want := map[string]int{
		"netconsoled_logs_received_total": 1,
		"netconsoled_logs_filter_total":   1,
		"netconsoled_logs_sink_total":     1,
	}

	for _, mf := range mfs {
		name := mf.GetName()
		value, ok := want[name]
		if !ok {
			continue
		}

		for _, m := range mf.GetMetric() {
			// TODO(mdlayher): expand to other types as needed.
			v := int(m.GetCounter().GetValue())

			if diff := cmp.Diff(value, v); diff != "" {
				t.Fatalf("unexpected metric %q value (-want +got):\n%s", name, diff)
			}
		}
	}
}
