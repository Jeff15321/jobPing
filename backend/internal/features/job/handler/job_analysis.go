package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jobping/backend/internal/features/job/service"
)

type JobAnalysisHandler struct {
	service *service.JobService
}

func NewJobAnalysisHandler(svc *service.JobService) *JobAnalysisHandler {
	return &JobAnalysisHandler{service: svc}
}

// HandleEvent processes SQS messages containing jobs for AI analysis and user matching
func (h *JobAnalysisHandler) HandleEvent(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		log.Printf("Processing job analysis message: %s", record.MessageId)

		var jobInput service.JobInput
		if err := json.Unmarshal([]byte(record.Body), &jobInput); err != nil {
			log.Printf("Failed to parse job message: %v", err)
			continue
		}

		job, err := h.service.ProcessJob(ctx, &jobInput)
		if err != nil {
			log.Printf("Failed to process job: %v", err)
			continue
		}

		if job != nil {
			log.Printf("Processed job: %s at %s (score: %v)", job.Title, job.Company, job.AIScore)
		}
	}

	return nil
}


