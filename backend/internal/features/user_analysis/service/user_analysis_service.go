package service

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/jobping/backend/internal/features/job/repository"
	usermodel "github.com/jobping/backend/internal/features/user/model"
	userrepo "github.com/jobping/backend/internal/features/user/repository"
)

type UserAnalysisService struct {
	jobRepo          repository.JobRepository
	userRepo         userrepo.UserRepository
	matchRepo        userrepo.UserJobMatchRepository
	aiClient         AIClient
	notificationQueueURL string
	sqsClient        *sqs.Client
}

func NewUserAnalysisService(jobRepo repository.JobRepository, userRepo userrepo.UserRepository, matchRepo userrepo.UserJobMatchRepository, aiClient AIClient) *UserAnalysisService {
	notificationQueueURL := os.Getenv("NOTIFICATION_QUEUE_URL")
	
	var sqsClient *sqs.Client
	if notificationQueueURL != "" {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err == nil {
			sqsClient = sqs.NewFromConfig(cfg)
		}
	}

	return &UserAnalysisService{
		jobRepo:              jobRepo,
		userRepo:             userRepo,
		matchRepo:            matchRepo,
		aiClient:             aiClient,
		notificationQueueURL: notificationQueueURL,
		sqsClient:            sqsClient,
	}
}

// AnalyzeUserMatch analyzes if a job matches a user and creates a match record if positive
func (s *UserAnalysisService) AnalyzeUserMatch(ctx context.Context, jobID, userID uuid.UUID) error {
	// Fetch job
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Job not found: %s", jobID)
		return nil
	}

	// Fetch user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		log.Printf("User not found: %s", userID)
		return nil
	}

	if user.AIPrompt == nil || *user.AIPrompt == "" {
		log.Printf("User %s has no AI prompt, skipping", userID)
		return nil
	}

	// Check if already matched
	existing, err := s.matchRepo.GetByUserAndJob(ctx, userID, jobID)
	if err != nil {
		log.Printf("Failed to check existing match: %v", err)
		return err
	}
	if existing != nil {
		log.Printf("Match already exists for user %s and job %s", userID, jobID)
		// Still enqueue to notification if not notified and score >= threshold
		if !existing.Notified && existing.Score >= user.NotifyThreshold {
			return s.enqueueToNotification(ctx, jobID, userID)
		}
		return nil
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
		return err
	}

	// Store match result
	match := &usermodel.UserJobMatch{
		ID:        uuid.New(),
		UserID:    userID,
		JobID:     jobID,
		Score:     matchResult.Score,
		Analysis:  matchResult.Analysis,
		Notified:  false,
		CreatedAt: time.Now(),
	}

	if err := s.matchRepo.Create(ctx, match); err != nil {
		log.Printf("Failed to save match: %v", err)
		return err
	}

	log.Printf("Matched job %s to user %s with score %d", job.Title, user.Username, matchResult.Score)

	// If match score >= threshold, enqueue to notification
	if matchResult.Score >= user.NotifyThreshold {
		return s.enqueueToNotification(ctx, jobID, userID)
	}

	return nil
}

func (s *UserAnalysisService) enqueueToNotification(ctx context.Context, jobID, userID uuid.UUID) error {
	if s.sqsClient == nil || s.notificationQueueURL == "" {
		log.Printf("SQS not configured, skipping notification enqueue")
		return nil
	}

	message := map[string]string{
		"job_id":  jobID.String(),
		"user_id": userID.String(),
	}
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = s.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.notificationQueueURL),
		MessageBody: aws.String(string(body)),
	})

	return err
}


