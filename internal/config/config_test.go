package config_test

import (
	"strings"
	"testing"

	"github.com/mdlayher/netconsoled"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/netconsoled/internal/config"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		cfg  *config.Config
		ok   bool
	}{
		{
			name: "empty server UDP",
			b: []byte(strings.TrimSpace(`
---
server:
			`)),
		},
		{
			name: "bad server UDP",
			b: []byte(strings.TrimSpace(`
---
server:
  udp_addr: :foo
			`)),
		},
		{
			name: "bad server HTTP",
			b: []byte(strings.TrimSpace(`
---
server:
  udp_addr: :6666
  http_addr: :foo
			`)),
		},
		{
			name: "bad filter",
			b: []byte(strings.TrimSpace(`
---
server:
  udp_addr: :6666
filters:
  - type: noop
  - type: bad
			`)),
		},
		{
			name: "bad sink",
			b: []byte(strings.TrimSpace(`
---
server:
  udp_addr: :6666
sinks:
  - type: noop
  - type: bad
			`)),
		},
		{
			name: "empty filters and sinks",
			b: []byte(strings.TrimSpace(`
---
server:
  udp_addr: :6666
  http_addr: :8080
			`)),
			cfg: &config.Config{
				Server: config.ServerConfig{
					UDPAddr:  ":6666",
					HTTPAddr: ":8080",
				},
				Filters: []netconsoled.Filter{
					netconsoled.NoopFilter(),
				},
				Sinks: []netconsoled.Sink{
					netconsoled.NoopSink(),
				},
			},
			ok: true,
		},
		{
			name: "empty filters",
			b: []byte(strings.TrimSpace(`
---
server:
  udp_addr: :6666
sinks:
  - type: noop
			`)),
			cfg: &config.Config{
				Server: config.ServerConfig{
					UDPAddr: ":6666",
				},
				Filters: []netconsoled.Filter{
					netconsoled.NoopFilter(),
				},
				Sinks: []netconsoled.Sink{
					netconsoled.NoopSink(),
				},
			},
			ok: true,
		},
		{
			name: "empty sinks",
			b: []byte(strings.TrimSpace(`
---
server:
  udp_addr: :6666
filters:
  - type: noop
			`)),
			cfg: &config.Config{
				Server: config.ServerConfig{
					UDPAddr: ":6666",
				},
				Filters: []netconsoled.Filter{
					netconsoled.NoopFilter(),
				},
				Sinks: []netconsoled.Sink{
					netconsoled.NoopSink(),
				},
			},
			ok: true,
		},
		{
			name: "multiple filters",
			b: []byte(strings.TrimSpace(`
---
server:
  udp_addr: :6666
filters:
  - type: noop
  - type: noop
sinks:
  - type: noop
			`)),
			cfg: &config.Config{
				Server: config.ServerConfig{
					UDPAddr: ":6666",
				},
				Filters: []netconsoled.Filter{
					netconsoled.NoopFilter(),
					netconsoled.NoopFilter(),
				},
				Sinks: []netconsoled.Sink{
					netconsoled.NoopSink(),
				},
			},
			ok: true,
		},
		{
			name: "multiple sinks",
			b: []byte(strings.TrimSpace(`
---
server:
  udp_addr: :6666
filters:
  - type: noop
sinks:
  - type: noop
  - type: noop
			`)),
			cfg: &config.Config{
				Server: config.ServerConfig{
					UDPAddr: ":6666",
				},
				Filters: []netconsoled.Filter{
					netconsoled.NoopFilter(),
				},
				Sinks: []netconsoled.Sink{
					netconsoled.NoopSink(),
					netconsoled.NoopSink(),
				},
			},
			ok: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.Parse(tt.b)

			if tt.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("expected an error, but none occurred: %v", err)
			}

			if !tt.ok {
				// Don't bother doing comparison if config is invalid.
				t.Logf("OK error: %v", err)
				return
			}

			if diff := cmp.Diff(tt.cfg, cfg); diff != "" {
				t.Fatalf("unexpected config (-want +got):\n%s", diff)
			}
		})
	}
}
