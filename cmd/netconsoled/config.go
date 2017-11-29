package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mdlayher/netconsoled/internal/config"
)

func initConfig(ll *log.Logger, file string) {
	ll.Printf("creating netconsoled configuration file %q", file)

	const defaultYAML = `---
# Configuration of the netconsoled server.
server:
  # Required: listen for incoming netconsole logs.
  udp_addr: :6666
  # Optional: enable HTTP server for Prometheus metrics.
  http_addr: :8080
# Zero or more filters to apply to incoming logs.
filters:
  # By default, apply no filtering to logs.
  - type: noop
# Zero or more sinks to use to store processed logs.
sinks:
  # By default, print logs to stdout and to a file.
  - type: stdout
  - type: file
    file: netconsoled.log
`

	if err := ioutil.WriteFile(file, []byte(defaultYAML), 0644); err != nil {
		ll.Fatalf("failed to write default configuration file %q: %v", file, err)
	}
}

func parseConfig(ll *log.Logger, file string) (*config.Config, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %v", err)
	}

	cfg, err := config.Parse(b)
	if err != nil {
		return nil, fmt.Errorf("failed to process configuration file: %v", err)
	}

	ll.Printf("loaded %d filter(s):", len(cfg.Filters))
	for _, f := range cfg.Filters {
		ll.Printf("  - %s", f.String())
	}

	ll.Printf("loaded %d sink(s):", len(cfg.Sinks))
	for _, s := range cfg.Sinks {
		ll.Printf("  - %s", s.String())
	}

	return cfg, nil
}
