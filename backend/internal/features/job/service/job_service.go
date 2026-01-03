package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jobping/backend/internal/features/job/model"
	"github.com/jobping/backend/internal/features/job/repository"
	usermodel "github.com/jobping/backend/internal/features/user/model"
	userrepo "github.com/jobping/backend/internal/features/user/repository"
)

type JobService struct {
	repo      repository.JobRepository
	aiClient  AIClient
	userRepo  userrepo.UserRepository
	matchRepo userrepo.UserJobMatchRepository
	sqsClient SQSClient
}

func NewJobService(repo repository.JobRepository, aiClient AIClient, userRepo userrepo.UserRepository, matchRepo userrepo.UserJobMatchRepository) *JobService {
	return &JobService{
		repo:      repo,
		aiClient:  aiClient,
		userRepo:  userRepo,
		matchRepo: matchRepo,
		sqsClient: NewSQSClient(),
	}
}

// ProcessJob receives a job from SQS, runs AI analysis, company research, and user matching
func (s *JobService) ProcessJob(ctx context.Context, input *JobInput) (*model.Job, error) {
	// Check if job already exists
	existingJob, err := s.repo.GetByURL(ctx, input.JobURL)
	if err != nil {
		return nil, err
	}
	if existingJob != nil {
		log.Printf("Job already exists: %s", input.JobURL)
		// Still run user matching for existing jobs if they haven't been matched yet
		if existingJob.Status == model.JobStatusProcessed {
			s.matchJobToUsers(ctx, existingJob)
		}
		return existingJob, nil
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

	// Step 1: Research company
	companyInfo, err := s.aiClient.ResearchCompany(ctx, job.Company, job.Title, job.Description)
	if err != nil {
		log.Printf("Company research failed: %v", err)
	} else {
		job.CompanyInfo = companyInfo
	}

	// Step 2: Run general AI analysis
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

	// Step 3: Match to all users with AI prompts
	if job.Status == model.JobStatusProcessed {
		s.matchJobToUsers(ctx, job)
	}

	return job, nil
}

// matchJobToUsers compares a job against all users with AI prompts and stores matches
func (s *JobService) matchJobToUsers(ctx context.Context, job *model.Job) {
	if s.userRepo == nil || s.matchRepo == nil {
		log.Printf("User/match repos not configured, skipping user matching")
		return
	}

	users, err := s.userRepo.GetUsersWithPrompts(ctx)
	if err != nil {
		log.Printf("Failed to get users with prompts: %v", err)
		return
	}

	log.Printf("Matching job %s to %d users", job.Title, len(users))

	for _, user := range users {
		if user.AIPrompt == nil || *user.AIPrompt == "" {
			continue
		}

		// Check if already matched
		existing, err := s.matchRepo.GetByUserAndJob(ctx, user.ID, job.ID)
		if err != nil {
			log.Printf("Failed to check existing match: %v", err)
			continue
		}
		if existing != nil {
			continue // Already matched
		}

		// Run AI matching
		matchInput := &JobMatchInput{
			Title:       job.Title,
			Company:     job.Company,
			Description: job.Description,
			CompanyInfo: job.CompanyInfo,
		}

		matchResult, err := s.aiClient.MatchJobToUser(ctx, matchInput, *user.AIPrompt)
		if err != nil {
			log.Printf("Failed to match job to user %s: %v", user.Username, err)
			continue
		}

		// Store match result
		match := &usermodel.UserJobMatch{
			ID:        uuid.New(),
			UserID:    user.ID,
			JobID:     job.ID,
			Score:     matchResult.Score,
			Analysis:  matchResult.Analysis,
			Notified:  false,
			CreatedAt: time.Now(),
		}

		if err := s.matchRepo.Create(ctx, match); err != nil {
			log.Printf("Failed to save match: %v", err)
			continue
		}

		log.Printf("Matched job %s to user %s with score %d", job.Title, user.Username, matchResult.Score)

		// Check if should notify
		if matchResult.Score >= user.NotifyThreshold && user.DiscordWebhook != nil {
			s.queueNotification(ctx, user, job, matchResult)
		}
	}
}

// queueNotification sends a notification request to the notification SQS queue
func (s *JobService) queueNotification(ctx context.Context, user usermodel.User, job *model.Job, match *UserMatchResult) {
	if user.DiscordWebhook == nil || *user.DiscordWebhook == "" {
		log.Printf("User %s has no Discord webhook configured, skipping notification", user.Username)
		return
	}

	notification := &NotificationMessage{
		DiscordWebhook: *user.DiscordWebhook,
		JobTitle:       job.Title,
		Company:        job.Company,
		JobURL:         job.JobURL,
		Score:          match.Score,
		Analysis:       match.Analysis,
	}

	if err := s.sqsClient.SendNotification(ctx, notification); err != nil {
		log.Printf("Failed to queue notification for user %s: %v", user.Username, err)
		return
	}

	log.Printf("Queued notification for user %s about job %s (score: %d)", user.Username, job.Title, match.Score)
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


