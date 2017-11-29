package netconsoled

import (
	"log"
	"net"
	"os"

	"github.com/mdlayher/netconsole"
	"github.com/prometheus/client_golang/prometheus"
)

// Data carries a netconsole log and its metadata.
type Data struct {
	Addr net.Addr
	Log  netconsole.Log
}

// A Server serves the netconsoled UDP and HTTP servers.
type Server struct {
	// Filter determines which received logs are allowed to be processed.
	Filter Filter

	// Sink gathers processed logs and stores them.
	Sink Sink

	// ErrorLog specifies a logger to use for capturing errors.
	ErrorLog *log.Logger

	// Metrics instruments a Server with Prometheus metrics, but only if
	// the Metrics structure is not empty.  This structure should be
	// populated using NewMetrics.
	Metrics

	// Embedding is used for metrics to slightly simplify the call sites
	// to s.inc and avoid polluting the Server structure, although this
	// probably isn't ideal.
}

// Prometheus metric labels.
const (
	labelOK      = "ok"
	labelDropped = "dropped"
	labelError   = "error"
)

// Handle handles incoming netconsole log messages.
func (s *Server) Handle(addr net.Addr, l netconsole.Log) {
	// Package up information for easier parameter passing.
	in := Data{
		Addr: addr,
		Log:  l,
	}

	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		s.ErrorLog.Printf("error splitting network address: %v", err)
		return
	}

	s.inc(s.LogsReceivedTotal, host)

	out, pass, err := s.Filter.Filter(in)
	if err != nil {
		s.inc(s.LogsFilterTotal, host, labelError)
		s.ErrorLog.Printf("error filtering log: %v", err)
		return
	}
	if !pass {
		s.inc(s.LogsFilterTotal, host, labelDropped)
		return
	}

	s.inc(s.LogsFilterTotal, host, labelOK)

	if err := s.Sink.Store(out); err != nil {
		s.inc(s.LogsSinkTotal, host, labelError)
		s.ErrorLog.Printf("error sending log to sink: %v", err)
		return
	}

	s.inc(s.LogsSinkTotal, host, labelOK)
}

// inc increments the specified counter with the specified labels.
// If metrics are not configured, inc is a no-op.
func (s *Server) inc(cv *prometheus.CounterVec, labels ...string) {
	if s.Metrics == (Metrics{}) || cv == nil {
		return
	}

	cv.WithLabelValues(labels...).Inc()
}

// Metrics contains Prometheus metrics for a Server.
type Metrics struct {
	// Metrics related to log ingestion and processing.
	LogsReceivedTotal *prometheus.CounterVec
	LogsFilterTotal   *prometheus.CounterVec
	LogsSinkTotal     *prometheus.CounterVec
}

// NewMetrics sets up a Metrics structure for a Server, and also returns
// a Prometheus registry which can be used to serve them.
func NewMetrics() (Metrics, *prometheus.Registry) {
	const (
		namespace    = "netconsoled"
		logSubsystem = "logs"

		labelHost   = "host"
		labelStatus = "status"
	)

	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(os.Getpid(), ""))

	logsRecv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: logSubsystem,
		Name:      "received_total",
		Help:      "Total number of logs received by the UDP server.",
	}, []string{labelHost})
	reg.MustRegister(logsRecv)

	logsFilter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: logSubsystem,
		Name:      "filter_total",
		Help:      "Total number of logs passed through a filter by status.",
	}, []string{labelHost, labelStatus})
	reg.MustRegister(logsFilter)

	logsSink := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: logSubsystem,
		Name:      "sink_total",
		Help:      "Total number of logs passed to a sink by status.",
	}, []string{labelHost, labelStatus})
	reg.MustRegister(logsSink)

	return Metrics{
		LogsReceivedTotal: logsRecv,
		LogsFilterTotal:   logsFilter,
		LogsSinkTotal:     logsSink,
	}, reg
}
