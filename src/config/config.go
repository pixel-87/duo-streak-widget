package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config keeps only the values we absolutely need right now. Add more fields
// when you start wiring the actual service.

	Port int
}

// LoadConfig reads the DUO_STREAK_PORT env var (optional) and falls back to 8080.
func LoadConfig() (Config, error) {
	cfg := Config{Port: 8080}
	if v := os.Getenv("DUO_STREAK_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p <= 0 || p > 65535 {
			return cfg, fmt.Errorf("invalid DUO_STREAK_PORT: %q", v)
		}
		cfg.Port = p
	}
	return cfg, nil
}
	// PORT (default 8080)
