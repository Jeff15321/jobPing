package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jobping/backend/internal/app"
	"github.com/jobping/backend/internal/sqspoller"
)

func main() {
	// Load environment variables
	queueURL := os.Getenv("NOTIFICATION_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("NOTIFICATION_QUEUE_URL environment variable is required")
	}

	// Build application
	appInstance, err := app.BuildNotifier()
	if err != nil {
		log.Fatalf("Failed to build application: %v", err)
	}

	// Create poller
	poller, err := sqspoller.New(queueURL, appInstance.SQSHandler.HandleSQSEvent)
	if err != nil {
		log.Fatalf("Failed to create SQS poller: %v", err)
	}

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping poller...")
		cancel()
	}()

	// Start polling
	log.Printf("Starting notifier worker for queue: %s", queueURL)
	if err := poller.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Poller error: %v", err)
	}

	log.Println("Notifier worker stopped")
}
