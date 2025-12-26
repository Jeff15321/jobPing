package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTExpiry   int // hours
}

func Load() *Config {
	// Load .env file in local development (ignore error if not found)
	_ = godotenv.Load()

	return &Config{
		Environment: getEnv("ENVIRONMENT", "local"),
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTExpiry:   getEnvInt("JWT_EXPIRY_HOURS", 24),
	}
}

func (c *Config) IsLocal() bool {
	return c.Environment == "local"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
