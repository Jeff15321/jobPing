package main

import (
	"log"
	"net/http"

	"github.com/jobping/backend/internal/app"
	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
)

func main() {
	cfg := config.Load()

	app, err := app.Build()
	if err != nil {
		log.Fatalf("Failed to build application: %v", err)
	}

	log.Println("Running database migrations...")
	if err := database.RunMigrations(cfg.DatabaseURL, "internal/database/migrations"); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	log.Printf("🚀 Server starting on http://localhost:%s", cfg.Port)
	log.Printf("📝 Environment: %s", cfg.Environment)

	if err := http.ListenAndServe(":"+cfg.Port, app.Router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
