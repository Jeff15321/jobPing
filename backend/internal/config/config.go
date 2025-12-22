package config

import (
	"os"
	"strconv"
)

type Config struct {
	Environment         string
	DatabaseURL         string
	AWSRegion           string
	AWSEndpoint         string
	SQSQueueURL         string
	SpeedyApplyAPIURL   string
	OpenAIAPIKey        string
	EmailFrom           string
	SESRegion           string
	APIPort             string
	FrontendURL         string
	ScanIntervalMinutes int
}

func Load() *Config {
	scanInterval, _ := strconv.Atoi(getEnv("SCAN_INTERVAL_MINUTES", "10"))

	return &Config{
		Environment:         getEnv("ENVIRONMENT", "local"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable"),
		AWSRegion:           getEnv("AWS_REGION", "us-east-1"),
		AWSEndpoint:         getEnv("AWS_ENDPOINT", ""),
		SQSQueueURL:         getEnv("SQS_QUEUE_URL", ""),
		SpeedyApplyAPIURL:   getEnv("SPEEDYAPPLY_API_URL", "https://api.speedyapply.com"),
		OpenAIAPIKey:        getEnv("OPENAI_API_KEY", ""),
		EmailFrom:           getEnv("EMAIL_FROM", "noreply@jobscanner.com"),
		SESRegion:           getEnv("SES_REGION", "us-east-1"),
		APIPort:             getEnv("API_PORT", "8080"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:5173"),
		ScanIntervalMinutes: scanInterval,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
