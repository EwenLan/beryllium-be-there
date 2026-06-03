package config

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// ServerConfig holds all server configuration.
type ServerConfig struct {
	Port     int    `toml:"port"`
	Hostname string `toml:"hostname"`
}

// DefaultConfig returns a ServerConfig with sensible defaults.
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Port: 8080,
	}
}

// Load reads config.toml from the given path and returns the merged configuration.
// Environment variables override file values: PORT, HOSTNAME.
func Load(path string) (ServerConfig, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist — use defaults
		} else {
			return cfg, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Environment variable overrides
	if p := os.Getenv("PORT"); p != "" {
		var port int
		if _, err := fmt.Sscanf(p, "%d", &port); err == nil {
			cfg.Port = port
		}
	}
	if h := os.Getenv("HOSTNAME"); h != "" {
		cfg.Hostname = h
	}

	// Auto-detect hostname if not configured
	if cfg.Hostname == "" {
		cfg.Hostname = detectLocalIP()
	}

	return cfg, nil
}

// detectLocalIP finds the first non-loopback IPv4 address on this machine.
func detectLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}

	for _, iface := range interfaces {
		// Skip down, loopback, and interfaces without multicast
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagMulticast == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip loopback, non-IPv4, link-local addresses
			if ip == nil || ip.IsLoopback() || ip.To4() == nil || ip.IsLinkLocalUnicast() {
				continue
			}

			return ip.String()
		}
	}

	// Fallback: try to get hostname
	hostname, err := os.Hostname()
	if err == nil && !strings.Contains(hostname, "localhost") {
		// Try to resolve the hostname to an IP
		if ips, err := net.LookupIP(hostname); err == nil {
			for _, ip := range ips {
				if ip.To4() != nil && !ip.IsLoopback() {
					return ip.String()
				}
			}
		}
	}

	return "localhost"
}
