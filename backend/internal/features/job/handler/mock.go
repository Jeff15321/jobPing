package handler

import (
	"net/http"

	"github.com/jobping/backend/internal/features/job/service"
)

// MockFetchJobs simulates the JobSpy Lambda for local development
// In production, this endpoint doesn't exist - the Python Lambda handles it
func (h *JobHandler) MockFetchJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Mock jobs that would come from JobSpy
	mockJobs := []service.JobInput{
		{
			Title:       "Senior Software Engineer",
			Company:     "TechCorp Inc",
			Location:    "San Francisco, CA",
			JobURL:      "https://example.com/job/1",
			Description: "We are looking for a senior software engineer with 5+ years of experience in Go, Python, or Java. You will be working on distributed systems and cloud infrastructure.",
			JobType:     "fulltime",
			IsRemote:    true,
		},
		{
			Title:       "Backend Developer",
			Company:     "StartupXYZ",
			Location:    "New York, NY",
			JobURL:      "https://example.com/job/2",
			Description: "Join our fast-growing startup as a backend developer. We use Go, PostgreSQL, and AWS. Great benefits and equity package.",
			JobType:     "fulltime",
			IsRemote:    false,
		},
		{
			Title:       "Full Stack Engineer",
			Company:     "BigTech Co",
			Location:    "Seattle, WA",
			JobURL:      "https://example.com/job/3",
			Description: "Looking for a full stack engineer proficient in React and Node.js. Experience with TypeScript and cloud services preferred.",
			JobType:     "fulltime",
			IsRemote:    true,
		},
	}

	processed := 0
	for _, jobInput := range mockJobs {
		job, err := h.service.ProcessJob(ctx, &jobInput)
		if err != nil {
			continue
		}
		if job != nil {
			processed++
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":     "Mock jobs fetched and processed",
		"jobs_found":  len(mockJobs),
		"jobs_queued": processed,
	})
}

