package service

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSClient interface for sending notifications
type SQSClient interface {
	SendNotification(ctx context.Context, notification *NotificationMessage) error
}

type NotificationMessage struct {
	DiscordWebhook string                 `json:"discord_webhook"`
	JobTitle       string                 `json:"job_title"`
	Company        string                 `json:"company"`
	JobURL         string                 `json:"job_url"`
	Score          int                    `json:"score"`
	Analysis       map[string]interface{} `json:"analysis"`
}

type sqsClient struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSClient() SQSClient {
	queueURL := os.Getenv("NOTIFY_SQS_QUEUE_URL")
	if queueURL == "" {
		// Return a no-op client if not configured
		return &noopSQSClient{}
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return &noopSQSClient{}
	}

	return &sqsClient{
		client:   sqs.NewFromConfig(cfg),
		queueURL: queueURL,
	}
}

func (c *sqsClient) SendNotification(ctx context.Context, notification *NotificationMessage) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	_, err = c.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(string(body)),
	})

	return err
}

// noopSQSClient is used when SQS is not configured (local development)
type noopSQSClient struct{}

func (c *noopSQSClient) SendNotification(ctx context.Context, notification *NotificationMessage) error {
	// No-op: just log in the calling code
	return nil
}

