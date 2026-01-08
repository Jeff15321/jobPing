package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/jobping/backend/internal/features/job_analysis/service"
)

type SQSHandler struct {
	service *service.JobAnalysisService
}

func NewSQSHandler(svc *service.JobAnalysisService) *SQSHandler {
	return &SQSHandler{service: svc}
}

// HandleSQSEvent processes SQS messages containing job_id
func (h *SQSHandler) HandleSQSEvent(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		log.Printf("Processing SQS message: %s", record.MessageId)

		var message struct {
			JobID string `json:"job_id"`
		}
		if err := json.Unmarshal([]byte(record.Body), &message); err != nil {
			log.Printf("Failed to parse SQS message: %v", err)
			return fmt.Errorf("failed to parse SQS message %s: %w", record.MessageId, err)
		}

		jobID, err := uuid.Parse(message.JobID)
		if err != nil {
			log.Printf("Invalid job_id in message: %v", err)
			return fmt.Errorf("invalid job_id in message %s: %w", record.MessageId, err)
		}

		if err := h.service.AnalyzeJob(ctx, jobID); err != nil {
			return fmt.Errorf("failed to analyze job %s: %w", jobID, err)
		}

		log.Printf("Processed job analysis for job_id: %s", jobID)
	}

	return nil
}

