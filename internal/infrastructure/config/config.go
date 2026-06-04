package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port string
	Env  string
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "taskdb"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

// DSN returns a PostgreSQL connection string.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

// getEnv returns the value of an environment variable or a default value if not set.
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}
