package main

import (
	"log"
	"net/http"

	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
	"github.com/jobping/backend/internal/server"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Running database migrations...")
	if err := database.RunMigrations(cfg.DatabaseURL, "internal/database/migrations"); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	router := server.NewRouter(cfg, db)

	log.Printf("🚀 Server starting on http://localhost:%s", cfg.Port)
	log.Printf("📝 Environment: %s", cfg.Environment)

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
