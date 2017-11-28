package netconsoled

import (
	"log"
	"net"
	"net/http"

	"github.com/mdlayher/netconsole"
)

// A Server serves the netconsoled UDP and HTTP servers.
type Server struct {
	// Filter determines which received logs are allowed to be processed.
	Filter Filter

	// Sink gathers processed logs and stores them.
	Sink Sink

	// ErrorLog specifies a logger to use for capturing errors.
	ErrorLog *log.Logger
}

// Handle handles incoming netconsole log messages.
func (s *Server) Handle(addr net.Addr, l *netconsole.Log) {
	// TODO(mdlayher): hooks/metrics in various areas.

	if !s.Filter.Allow(addr, l) {
		return
	}

	if err := s.Sink.Store(addr, l); err != nil {
		s.ErrorLog.Printf("error sending log to sink: %v", err)
		return
	}
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	return
}
