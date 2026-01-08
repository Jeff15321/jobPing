package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/jobping/backend/internal/features/user_fanout/service"
)

type SQSHandler struct {
	service *service.FanoutService
}

func NewSQSHandler(svc *service.FanoutService) *SQSHandler {
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
			continue
		}

		jobID, err := uuid.Parse(message.JobID)
		if err != nil {
			log.Printf("Invalid job_id in message: %v", err)
			continue
		}

		if err := h.service.FanoutToUsers(ctx, jobID); err != nil {
			log.Printf("Failed to fanout to users: %v", err)
			continue
		}

		log.Printf("Processed fanout for job_id: %s", jobID)
	}

	return nil
}


