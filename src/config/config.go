package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	Port                  int
	CacheTTL              time.Duration
	FallbackTTL           time.Duration
	DuolingoBaseURL       string
	DuolingoTimeout       time.Duration
	RateLimitPerMin       int
	MaxConcurrentRequests int
	LogLevel              string
}

// LoadConfig reads environment variables and returns a validated Config.
func LoadConfig() (Config, error) {
	var cfg Config

	// PORT (default 8080)
	if v := os.Getenv("DUO_STREAK_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p <= 0 || p > 65535 {
			return cfg, fmt.Errorf("invalid DUO_STREAK_PORT: %q", v)
		}
		cfg.Port = p
	} else {
		cfg.Port = 8080
	}

	// Cache TTL (default 2h)
	if v := os.Getenv("DUO_STREAK_CACHE_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return cfg, fmt.Errorf("invalid DUO_STREAK_CACHE_TTL: %q", v)
		}
		cfg.CacheTTL = d
	} else {
		cfg.CacheTTL = 2 * time.Hour
	}

	// Fallback TTL (default 5m)
	if v := os.Getenv("DUO_STREAK_FALLBACK_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return cfg, fmt.Errorf("invalid DUO_STREAK_FALLBACK_TTL: %q", v)
		}
		cfg.FallbackTTL = d
	} else {
		cfg.FallbackTTL = 5 * time.Minute
	}

	// Duolingo base URL (default https://www.duolingo.com)
	if v := os.Getenv("DUO_STREAK_DUOLINGO_BASE_URL"); v != "" {
		u, err := url.ParseRequestURI(v)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
			return cfg, fmt.Errorf("invalid DUO_STREAK_DUOLINGO_BASE_URL: %q", v)
		}
		cfg.DuolingoBaseURL = v
	} else {
		cfg.DuolingoBaseURL = "https://www.duolingo.com"
	}

	// Duolingo timeout (default 10s)
	if v := os.Getenv("DUO_STREAK_DUOLINGO_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return cfg, fmt.Errorf("invalid DUO_STREAK_DUOLINGO_TIMEOUT: %q", v)
		}
		cfg.DuolingoTimeout = d
	} else {
		cfg.DuolingoTimeout = 10 * time.Second
	}

	// RateLimit per minute (default 60)
	if v := os.Getenv("DUO_STREAK_RATE_LIMIT_PER_MIN"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return cfg, fmt.Errorf("invalid DUO_STREAK_RATE_LIMIT_PER_MIN: %q", v)
		}
		cfg.RateLimitPerMin = n
	} else {
		cfg.RateLimitPerMin = 60
	}

	// Max concurrent outbound requests (default 5)
	if v := os.Getenv("DUO_STREAK_MAX_CONCURRENT_REQUESTS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return cfg, fmt.Errorf("invalid DUO_STREAK_MAX_CONCURRENT_REQUESTS: %q", v)
		}
		cfg.MaxConcurrentRequests = n
	} else {
		cfg.MaxConcurrentRequests = 5
	}

	// Log level
	if v := os.Getenv("DUO_STREAK_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	} else {
		cfg.LogLevel = "info"
	}

	return cfg, nil
}
