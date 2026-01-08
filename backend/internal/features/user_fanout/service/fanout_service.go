package service

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	userrepo "github.com/jobping/backend/internal/features/user/repository"
)

type FanoutService struct {
	userRepo         userrepo.UserRepository
	analysisQueueURL string
	sqsClient        *sqs.Client
}

func NewFanoutService(userRepo userrepo.UserRepository) *FanoutService {
	analysisQueueURL := os.Getenv("USER_ANALYSIS_QUEUE_URL")
	
	var sqsClient *sqs.Client
	if analysisQueueURL != "" {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err == nil {
			sqsClient = sqs.NewFromConfig(cfg)
		}
	}

	return &FanoutService{
		userRepo:         userRepo,
		analysisQueueURL: analysisQueueURL,
		sqsClient:        sqsClient,
	}
}

// FanoutToUsers fetches all users with AI prompts and enqueues job_id+user_id to user-analysis-queue
func (s *FanoutService) FanoutToUsers(ctx context.Context, jobID uuid.UUID) error {
	// Fetch all users with AI prompts
	users, err := s.userRepo.GetUsersWithPrompts(ctx)
	if err != nil {
		return err
	}

	log.Printf("Fanning out job %s to %d users", jobID, len(users))

	// Enqueue each user to user-analysis-queue
	enqueued := 0
	for _, user := range users {
		if user.AIPrompt == nil || *user.AIPrompt == "" {
			continue
		}

		if err := s.enqueueToAnalysis(ctx, jobID, user.ID); err != nil {
			log.Printf("Failed to enqueue user %s: %v", user.ID, err)
			continue
		}
		enqueued++
	}

	log.Printf("Enqueued job %s to %d users", jobID, enqueued)
	return nil
}

func (s *FanoutService) enqueueToAnalysis(ctx context.Context, jobID, userID uuid.UUID) error {
	if s.sqsClient == nil || s.analysisQueueURL == "" {
		log.Printf("SQS not configured, skipping analysis enqueue")
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
		QueueUrl:    aws.String(s.analysisQueueURL),
		MessageBody: aws.String(string(body)),
	})

	return err
}


