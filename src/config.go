package config

import "time"

type Config struct {
	Port     int
	CacheTTL time.Duration
}