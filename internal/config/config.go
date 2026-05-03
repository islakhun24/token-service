package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration.
type Config struct {
	Port                string
	CoinGeckoListTTL    time.Duration
	CoinGeckoMarketsTTL time.Duration
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBSSLMode           string
	CronSchedule        string
}

// DatabaseDSN constructs the database connection string.
func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode)
}

// Load loads configuration from environment variables.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	return &Config{
		Port:         getEnv("PORT", "8081"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", "postgres"),
		DBName:       getEnv("DB_NAME", "token_service"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		CronSchedule: getEnv("CRON_SCHEDULE", "@every 1h"),
	}
}

// getEnv fetches the environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
