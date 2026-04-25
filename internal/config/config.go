package config

import (
	"os"
	"time"
)

// Config holds application configuration.
type Config struct {
	Port                string
	CoinGeckoListTTL    time.Duration
	CoinGeckoMarketsTTL time.Duration
}

// Load loads configuration from environment variables.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	return &Config{
		Port:                port,
		CoinGeckoListTTL:    6 * time.Hour,
		CoinGeckoMarketsTTL: 60 * time.Second,
	}
}
