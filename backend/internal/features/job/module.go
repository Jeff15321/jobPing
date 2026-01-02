package job

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jobping/backend/internal/features/job/handler"
)

// RegisterRoutes registers job-related HTTP routes
func RegisterRoutes(r chi.Router, jobHandler *handler.JobHandler) {
	r.Get("/jobs", jobHandler.GetJobs)

	// Local development endpoints only
	// In production, these go through Python Lambda + SQS
	if os.Getenv("ENVIRONMENT") != "production" {
		r.Post("/jobs/fetch", jobHandler.MockFetchJobs)   // Mock fetch with fake jobs
		r.Post("/jobs/process", jobHandler.ProcessJob)    // Process real jobs from Python
	}
}

