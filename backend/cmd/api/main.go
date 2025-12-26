package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
	"github.com/jobping/backend/internal/server"
)

var chiLambda *chiadapter.ChiLambda

func init() {
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations in production (optional, can be run separately)
	if cfg.IsProduction() {
		log.Println("Running database migrations...")
		if err := database.RunMigrations(cfg.DatabaseURL, "internal/database/migrations"); err != nil {
			log.Printf("Migration warning: %v", err)
		}
	}

	// Create router
	router := server.NewRouter(cfg, db)

	// Setup Lambda adapter if running in Lambda
	if !cfg.IsLocal() {
		chiLambda = chiadapter.New(router)
	}
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return chiLambda.ProxyWithContext(ctx, req)
}

func main() {
	cfg := config.Load()

	if cfg.IsLocal() {
		// Local development mode
		db, err := database.Connect(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Run migrations locally
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
	} else {
		// AWS Lambda mode
		if os.Getenv("_LAMBDA_SERVER_PORT") != "" {
			lambda.Start(handler)
		} else {
			// Fallback to Lambda
			lambda.Start(handler)
		}
	}
}

