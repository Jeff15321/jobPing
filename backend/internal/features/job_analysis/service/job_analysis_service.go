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
	jobrepo "github.com/jobping/backend/internal/features/job/repository"
)

type JobAnalysisService struct {
	jobRepo      jobrepo.JobRepository
	aiClient     AIClient
	fanoutQueueURL string
	sqsClient    *sqs.Client
}

func NewJobAnalysisService(jobRepo jobrepo.JobRepository, aiClient AIClient) *JobAnalysisService {
	fanoutQueueURL := os.Getenv("USER_FANOUT_QUEUE_URL")
	
	var sqsClient *sqs.Client
	if fanoutQueueURL != "" {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err == nil {
			sqsClient = sqs.NewFromConfig(cfg)
		}
	}

	return &JobAnalysisService{
		jobRepo:        jobRepo,
		aiClient:       aiClient,
		fanoutQueueURL: fanoutQueueURL,
		sqsClient:      sqsClient,
	}
}

// AnalyzeJob processes a job by checking if company info is fresh, and if not, researching the company
func (s *JobAnalysisService) AnalyzeJob(ctx context.Context, jobID uuid.UUID) error {
	// Fetch job
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Job not found: %s", jobID)
		return nil
	}

	// Check if company info is fresh (< 6 months old)
	isFresh, err := s.jobRepo.IsCompanyInfoFresh(ctx, jobID)
	if err != nil {
		log.Printf("Failed to check company info freshness: %v", err)
		// Continue anyway
	}

	if isFresh && job.CompanyInfo != nil {
		log.Printf("Company info is fresh for job %s, skipping analysis", jobID)
	} else {
		// Research company using ChatGPT
		log.Printf("Researching company for job %s: %s", jobID, job.Company)
		companyInfo, err := s.aiClient.ResearchCompany(ctx, job.Company, job.Title, job.Description)
		if err != nil {
			return err
		} else {
			// Save company info
			if err := s.jobRepo.UpdateCompanyInfo(ctx, jobID, companyInfo); err != nil {
				log.Printf("Failed to save company info: %v", err)
			} else {
				log.Printf("Saved company info for job %s", jobID)
			}
		}
	}

	// Enqueue to user-fanout-queue
	if err := s.enqueueToFanout(ctx, jobID); err != nil {
		log.Printf("Failed to enqueue to fanout queue: %v", err)
		return err
	}

	return nil
}

func (s *JobAnalysisService) enqueueToFanout(ctx context.Context, jobID uuid.UUID) error {
	if s.sqsClient == nil || s.fanoutQueueURL == "" {
		log.Printf("SQS not configured, skipping fanout enqueue")
		return nil
	}

	message := map[string]string{
		"job_id": jobID.String(),
	}
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = s.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.fanoutQueueURL),
		MessageBody: aws.String(string(body)),
	})

	return err
}

