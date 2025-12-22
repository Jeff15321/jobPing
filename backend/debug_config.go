package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/yourusername/ai-job-scanner/internal/config"
)

func main() {
	// Try to load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	} else {
		log.Println(".env file loaded successfully")
	}

	// Load configuration
	cfg := config.Load()

	fmt.Printf("Environment: %s\n", cfg.Environment)
	fmt.Printf("Database URL: %s\n", cfg.DatabaseURL)
	fmt.Printf("API Port: %s\n", cfg.APIPort)
}