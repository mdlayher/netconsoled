package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mdlayher/netconsole"
	"github.com/mdlayher/netconsoled"
	"github.com/mdlayher/netconsoled/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func serve(ctx context.Context, ll *log.Logger, cfg *config.Config) {
	// Set up Prometheus metrics.
	metrics, reg := netconsoled.NewMetrics()

	// Sink is split out so it can shut down gracefully later.
	sink := netconsoled.MultiSink(cfg.Sinks...)

	s := &netconsoled.Server{
		Filter:   netconsoled.MultiFilter(cfg.Filters...),
		Sink:     sink,
		ErrorLog: ll,
		Metrics:  metrics,
	}

	// Start each network service in its own goroutine so they can
	// be shut down at a later time.
	var wg sync.WaitGroup

	// UDP server goroutine.
	wg.Add(1)
	go func() {
		defer wg.Done()

		ns := netconsole.NewServer("udp", cfg.Server.UDPAddr, s.Handle)

		ll.Printf("starting UDP server at %q", cfg.Server.UDPAddr)

		// Canceled context will stop listener.
		if err := ns.ListenAndServe(ctx); err != nil {
			ll.Fatalf("failed to listen UDP: %v", err)
		}
	}()

	// HTTP server goroutine, if enabled.
	if cfg.Server.HTTPAddr != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Blocks until stopped via context.
			serveHTTP(ctx, cfg.Server.HTTPAddr, reg, ll)
		}()
	}

	// Block main goroutine until all servers halt.
	wg.Wait()

	// If possible, flush sink data before shutdown.
	if c, ok := sink.(io.Closer); ok {
		if err := c.Close(); err != nil {
			ll.Fatalf("failed to flush sink data: %v", err)
		}

		ll.Println("flushed all sink data")
	}
}

func serveHTTP(ctx context.Context, addr string, reg *prometheus.Registry, ll *log.Logger) {
	// Set up Prometheus and future API.
	prom := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog: ll,
	})

	mux := http.NewServeMux()
	mux.Handle("/metrics", prom)

	hs := &http.Server{
		Addr:     addr,
		Handler:  mux,
		ErrorLog: ll,
	}

	// Block until both HTTP server goroutines are shut down.
	var wg sync.WaitGroup
	wg.Add(2)

	// HTTP server listener goroutine.
	go func() {
		defer wg.Done()

		ll.Printf("starting HTTP server at %q", addr)

		// Canceled via Shutdown.
		if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ll.Fatalf("failed to listen HTTP: %v", err)
		}
	}()

	// HTTP server shutdown goroutine.
	go func() {
		defer wg.Done()

		// Wait for the parent context to be done before shutting down.
		<-ctx.Done()

		// Parent context is already closed so start a new background context
		// for cancelation.
		hctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := hs.Shutdown(hctx); err != nil {
			ll.Fatalf("failed to shut down HTTP server: %v", err)
		}
	}()

	wg.Wait()
}
