package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/jobping/backend/internal/features/notification/service"
)

type SQSHandler struct {
	service *service.NotificationService
}

func NewSQSHandler(svc *service.NotificationService) *SQSHandler {
	return &SQSHandler{service: svc}
}

// HandleSQSEvent processes SQS messages containing job_id and user_id
func (h *SQSHandler) HandleSQSEvent(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		log.Printf("Processing SQS message: %s", record.MessageId)

		var message struct {
			JobID  string `json:"job_id"`
			UserID string `json:"user_id"`
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

		userID, err := uuid.Parse(message.UserID)
		if err != nil {
			log.Printf("Invalid user_id in message: %v", err)
			continue
		}

		if err := h.service.SendNotification(ctx, jobID, userID); err != nil {
			log.Printf("Failed to send notification: %v", err)
			continue
		}

		log.Printf("Processed notification for job_id: %s, user_id: %s", jobID, userID)
	}

	return nil
}


