package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jobping/backend/internal/features/job/service"
)

type SQSHandler struct {
	service *service.JobService
}

func NewSQSHandler(svc *service.JobService) *SQSHandler {
	return &SQSHandler{service: svc}
}

// HandleSQSEvent processes SQS messages containing jobs
func (h *SQSHandler) HandleSQSEvent(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		log.Printf("Processing SQS message: %s", record.MessageId)

		var jobInput service.JobInput
		if err := json.Unmarshal([]byte(record.Body), &jobInput); err != nil {
			log.Printf("Failed to parse SQS message: %v", err)
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

