package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/yourusername/ai-job-scanner/internal/config"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	log.Printf("Matcher service starting (environment: %s)", cfg.Environment)
	
	// TODO: Implement SQS queue listener and job matching logic
	// This will be implemented in the next phase
	
	log.Println("Matcher service ready")
	select {} // Keep running
}
