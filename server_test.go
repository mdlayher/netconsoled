package netconsoled_test

import (
	"net"
	"testing"

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
			name:   "filter disallow",
			verify: testServerFilterDisallow,
		},
		{
			name:   "sink ok",
			verify: testServerSinkOK,
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
		Filter: netconsoled.FuncFilter(func(d netconsoled.Data) bool {
			return false
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
