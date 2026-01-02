package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jobping/backend/internal/features/job/model"
	"github.com/jobping/backend/internal/features/job/repository"
)

type JobService struct {
	repo     repository.JobRepository
	aiClient AIClient
}

func NewJobService(repo repository.JobRepository, aiClient AIClient) *JobService {
	return &JobService{
		repo:     repo,
		aiClient: aiClient,
	}
}

// ProcessJob receives a job from SQS, runs AI analysis, and saves to DB
func (s *JobService) ProcessJob(ctx context.Context, input *JobInput) (*model.Job, error) {
	// Check if job already exists
	exists, err := s.repo.ExistsByURL(ctx, input.JobURL)
	if err != nil {
		return nil, err
	}
	if exists {
		log.Printf("Job already exists: %s", input.JobURL)
		return nil, nil
	}

	now := time.Now()
	job := &model.Job{
		ID:          uuid.New(),
		Title:       input.Title,
		Company:     input.Company,
		Location:    input.Location,
		JobURL:      input.JobURL,
		Description: input.Description,
		JobType:     input.JobType,
		IsRemote:    input.IsRemote,
		MinSalary:   input.MinSalary,
		MaxSalary:   input.MaxSalary,
		DatePosted:  input.DatePosted,
		Status:      model.JobStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Run AI analysis
	result, err := s.aiClient.AnalyzeJob(ctx, job.Title, job.Company, job.Description)
	if err != nil {
		log.Printf("AI analysis failed: %v", err)
		job.Status = model.JobStatusFailed
	} else {
		job.AIScore = &result.Score
		job.AIAnalysis = &result.Analysis
		job.Status = model.JobStatusProcessed
	}

	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

// GetJobs returns processed jobs for display
func (s *JobService) GetJobs(ctx context.Context, limit int) ([]model.Job, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.GetProcessed(ctx, limit)
}

// JobInput represents a job from SQS message
type JobInput struct {
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	Location    string   `json:"location"`
	JobURL      string   `json:"job_url"`
	Description string   `json:"description"`
	JobType     string   `json:"job_type"`
	IsRemote    bool     `json:"is_remote"`
	MinSalary   *float64 `json:"min_amount"`
	MaxSalary   *float64 `json:"max_amount"`
	DatePosted  string   `json:"date_posted"`
}

