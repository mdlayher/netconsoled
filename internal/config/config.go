package config

import (
	"errors"
	"fmt"
	"net"

	"github.com/mdlayher/netconsoled"
	yaml "gopkg.in/yaml.v2"
)

// Parse parses a Config from its raw YAML format.
func Parse(b []byte) (*Config, error) {
	var c RawConfig
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}

	if err := checkServerConfig(c.Server); err != nil {
		return nil, err
	}

	filters, err := parseFilters(c)
	if err != nil {
		return nil, err
	}

	sinks, err := parseSinks(c)
	if err != nil {
		return nil, err
	}

	return &Config{
		Server:  c.Server,
		Filters: filters,
		Sinks:   sinks,
	}, nil
}

// checkServerConfig validates a ServerConfig.
func checkServerConfig(c ServerConfig) error {
	if c.UDPAddr == "" {
		return errors.New("server UDP address must not be empty")
	}

	if _, err := net.ResolveUDPAddr("udp", c.UDPAddr); err != nil {
		return fmt.Errorf("failed to parse server UDP address: %v", err)
	}

	if c.HTTPAddr != "" {
		if _, err := net.ResolveTCPAddr("tcp", c.HTTPAddr); err != nil {
			return fmt.Errorf("failed to parse server HTTP address: %v", err)
		}
	}

	return nil
}

// parseFilters builds a slice of netconsoled.Filters from a RawConfig.
func parseFilters(c RawConfig) ([]netconsoled.Filter, error) {
	var fs []netconsoled.Filter
	for _, f := range c.Filters {
		var filter netconsoled.Filter

		switch f.Type {
		case "noop":
			filter = netconsoled.NoopFilter()
		default:
			return nil, fmt.Errorf("unknown filter type in configuration: %q", f.Type)
		}

		fs = append(fs, filter)
	}

	if len(fs) == 0 {
		fs = append(fs, netconsoled.NoopFilter())
	}

	return fs, nil
}

// parseSinks builds a slice of netconsoled.Sinks from a RawConfig.
func parseSinks(c RawConfig) ([]netconsoled.Sink, error) {
	var ss []netconsoled.Sink
	for _, s := range c.Sinks {
		var (
			sink netconsoled.Sink
			err  error
		)

		switch s.Type {
		case "file":
			if s.File == "" {
				return nil, errors.New("must specify output file for file sink")
			}

			sink, err = netconsoled.FileSink(s.File)
		case "noop":
			sink = netconsoled.NoopSink()
		case "stdout":
			sink = netconsoled.StdoutSink()
		default:
			return nil, fmt.Errorf("unknown sink type in configuration: %q", s.Type)
		}
		if err != nil {
			return nil, err
		}

		ss = append(ss, sink)
	}

	if len(ss) == 0 {
		ss = append(ss, netconsoled.NoopSink())
	}

	return ss, nil
}

// A RawConfig is the raw structure used to unmarshal YAML configuration.
type RawConfig struct {
	Server ServerConfig `yaml:"server"`

	Filters []struct {
		Type string `yaml:"type"`
	} `yaml:"filters"`

	Sinks []struct {
		Type string `yaml:"type"`
		File string `yaml:"file"`
	} `yaml:"sinks"`
}

// A Config is the processed configuration for a netconsoled server.
type Config struct {
	Server  ServerConfig
	Filters []netconsoled.Filter
	Sinks   []netconsoled.Sink
}

// A ServerConfig contains configuration for a netconsoled server's
// network listeners.
type ServerConfig struct {
	UDPAddr  string `yaml:"udp_addr"`
	HTTPAddr string `yaml:"http_addr"`
}
