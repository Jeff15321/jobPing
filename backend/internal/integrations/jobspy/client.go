package jobspy

import (
	"context"
	"net/http"
	"time"

	"github.com/yourusername/ai-job-scanner/internal/domain/job"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type SpeedyApplyJob struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PostedDate  string `json:"posted_date"`
}

type SpeedyApplyResponse struct {
	Jobs []SpeedyApplyJob `json:"jobs"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) FetchLatestJobs(ctx context.Context, limit int) ([]*job.Job, error) {
	// TODO: Replace with actual SpeedyApply API call
	// For now, return mock data for testing
	
	mockJobs := []*job.Job{
		{
			ID:          "mock-1",
			Title:       "Software Engineer Intern",
			Company:     "Google",
			Location:    "Mountain View, CA",
			Description: "Join our team as a Software Engineer Intern! You'll work on cutting-edge projects, collaborate with world-class engineers, and contribute to products used by billions of people worldwide. This internship offers hands-on experience with large-scale systems, machine learning, and innovative technologies.",
			URL:         "https://careers.google.com/jobs/mock-1",
			PostedAt:    time.Now().Add(-2 * time.Hour),
			FetchedAt:   time.Now(),
		},
		{
			ID:          "mock-2",
			Title:       "Frontend Developer",
			Company:     "Meta",
			Location:    "Menlo Park, CA",
			Description: "Build the future of social technology! As a Frontend Developer, you'll create engaging user experiences for billions of users. Work with React, JavaScript, and cutting-edge web technologies while collaborating with designers and product managers.",
			URL:         "https://careers.meta.com/jobs/mock-2",
			PostedAt:    time.Now().Add(-4 * time.Hour),
			FetchedAt:   time.Now(),
		},
		{
			ID:          "mock-3",
			Title:       "Backend Engineer",
			Company:     "Netflix",
			Location:    "Los Gatos, CA",
			Description: "Help us deliver entertainment to 200+ million members worldwide! You'll build scalable microservices, work with big data systems, and optimize performance for global streaming. Experience with Java, Python, or Go preferred.",
			URL:         "https://jobs.netflix.com/jobs/mock-3",
			PostedAt:    time.Now().Add(-6 * time.Hour),
			FetchedAt:   time.Now(),
		},
		{
			ID:          "mock-4",
			Title:       "Full Stack Developer",
			Company:     "Stripe",
			Location:    "San Francisco, CA",
			Description: "Build the economic infrastructure for the internet! Work on payment systems that process billions of dollars. You'll use Ruby, JavaScript, and modern frameworks to create developer-friendly APIs and beautiful user interfaces.",
			URL:         "https://stripe.com/jobs/mock-4",
			PostedAt:    time.Now().Add(-8 * time.Hour),
			FetchedAt:   time.Now(),
		},
		{
			ID:          "mock-5",
			Title:       "DevOps Engineer",
			Company:     "Airbnb",
			Location:    "San Francisco, CA",
			Description: "Scale our platform to serve millions of travelers and hosts! You'll work with Kubernetes, AWS, and infrastructure as code. Help us maintain 99.9% uptime while supporting rapid feature development.",
			URL:         "https://careers.airbnb.com/jobs/mock-5",
			PostedAt:    time.Now().Add(-12 * time.Hour),
			FetchedAt:   time.Now(),
		},
	}

	// Return only the requested number of jobs
	if limit > len(mockJobs) {
		limit = len(mockJobs)
	}

	return mockJobs[:limit], nil

	// Real API implementation would be:
	// url := fmt.Sprintf("%s/jobs?limit=%d&type=swe-internship", c.baseURL, limit)
	// ... rest of the HTTP request logic
}
