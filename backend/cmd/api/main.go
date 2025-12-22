package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/yourusername/ai-job-scanner/internal/api/handlers"
	"github.com/yourusername/ai-job-scanner/internal/api/middleware"
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

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize handlers
	fetcher := jobspy.NewClient(cfg.SpeedyApplyAPIURL)
	jobHandler := handlers.NewJobHandler(db, fetcher)
	userHandler := handlers.NewUserHandler(db)

	// Setup router
	r := mux.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	
	// Jobs
	api.HandleFunc("/jobs", jobHandler.GetJobs).Methods("GET")
	api.HandleFunc("/jobs/{id}", jobHandler.GetJob).Methods("GET")
	api.HandleFunc("/jobs/scan", jobHandler.ScanJobs).Methods("POST")
	
	// Users
	api.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id}/preferences", userHandler.UpdatePreferences).Methods("PUT")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL, "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	// Start server
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
