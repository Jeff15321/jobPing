package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/yourusername/ai-job-scanner/internal/config"
	"github.com/yourusername/ai-job-scanner/internal/database"
	"github.com/yourusername/ai-job-scanner/internal/integrations/jobspy"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize job fetcher
	fetcher := jobspy.NewClient(cfg.SpeedyApplyAPIURL)

	// Run scanner
	log.Println("Starting job scanner...")
	ctx := context.Background()

	// For Lambda: run once and exit
	// For local/cron: run in loop
	if cfg.Environment == "lambda" {
		if err := scanJobs(ctx, fetcher, db); err != nil {
			log.Fatalf("Scan failed: %v", err)
		}
	} else {
		// Local development: run every 10 minutes
		ticker := time.NewTicker(time.Duration(cfg.ScanIntervalMinutes) * time.Minute)
		defer ticker.Stop()

		// Run immediately on start
		if err := scanJobs(ctx, fetcher, db); err != nil {
			log.Printf("Scan failed: %v", err)
		}

		for range ticker.C {
			if err := scanJobs(ctx, fetcher, db); err != nil {
				log.Printf("Scan failed: %v", err)
			}
		}
	}
}

func scanJobs(ctx context.Context, fetcher *jobspy.Client, db *database.DB) error {
	log.Println("Fetching jobs from SpeedyApply...")

	jobs, err := fetcher.FetchLatestJobs(ctx, 10)
	if err != nil {
		return err
	}

	log.Printf("Fetched %d jobs", len(jobs))

	// Store jobs in database
	for _, job := range jobs {
		if err := database.UpsertJob(db, job); err != nil {
			log.Printf("Failed to store job %s: %v", job.ID, err)
			continue
		}
		log.Printf("Stored job: %s - %s", job.ID, job.Title)
	}

	return nil
}
