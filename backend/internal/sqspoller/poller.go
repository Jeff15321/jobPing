package sqspoller

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Poller polls an SQS queue and invokes a handler function for each message
type Poller struct {
	QueueURL    string
	Handler     func(ctx context.Context, event events.SQSEvent) error
	PollInterval time.Duration
	sqsClient   *sqs.Client
}

// New creates a new SQS poller
func New(queueURL string, handler func(ctx context.Context, event events.SQSEvent) error) (*Poller, error) {
	// Configure AWS SDK for LocalStack or real AWS
	endpointURL := os.Getenv("AWS_ENDPOINT_URL")
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Override endpoint for LocalStack
	if endpointURL != "" {
		cfg.BaseEndpoint = aws.String(endpointURL)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	return &Poller{
		QueueURL:     queueURL,
		Handler:      handler,
		PollInterval: 5 * time.Second,
		sqsClient:    sqsClient,
	}, nil
}

// Start begins polling the queue and processing messages
func (p *Poller) Start(ctx context.Context) error {
	log.Printf("Starting SQS poller for queue: %s", p.QueueURL)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping SQS poller for queue: %s", p.QueueURL)
			return ctx.Err()
		default:
			if err := p.pollAndProcess(ctx); err != nil {
				log.Printf("Error polling queue: %v", err)
			}
			time.Sleep(p.PollInterval)
		}
	}
}

// pollAndProcess performs a single poll operation
func (p *Poller) pollAndProcess(ctx context.Context) error {
	// Long poll (20 seconds) to reduce API calls
	result, err := p.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(p.QueueURL),
		MaxNumberOfMessages: 1, // Process one at a time
		WaitTimeSeconds:     20, // Long polling
		VisibilityTimeout:   60, // Message invisible for 60s while processing
	})
	if err != nil {
		return fmt.Errorf("failed to receive messages: %w", err)
	}

	if len(result.Messages) == 0 {
		return nil // No messages, continue polling
	}

	// Process each message
	for _, msg := range result.Messages {
		if err := p.processMessage(ctx, msg); err != nil {
			log.Printf("Failed to process message %s: %v", *msg.MessageId, err)
			// Don't delete message on error - let visibility timeout handle retry
			continue
		}

		// Delete message on success
		if _, err := p.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(p.QueueURL),
			ReceiptHandle: msg.ReceiptHandle,
		}); err != nil {
			log.Printf("Failed to delete message %s: %v", *msg.MessageId, err)
		}
	}

	return nil
}

// processMessage converts an SQS message to events.SQSEvent and calls the handler
func (p *Poller) processMessage(ctx context.Context, msg types.Message) error {
	// Construct events.SQSEvent from the message
	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId:              *msg.MessageId,
				ReceiptHandle:          *msg.ReceiptHandle,
				Body:                   *msg.Body,
				Md5OfBody:              aws.ToString(msg.MD5OfBody),
				Md5OfMessageAttributes: aws.ToString(msg.MD5OfMessageAttributes),
				Attributes:              msg.Attributes,
				MessageAttributes:      convertMessageAttributes(msg.MessageAttributes),
				EventSourceARN:         msg.Attributes["QueueArn"],
				EventSource:             "aws:sqs",
				AWSRegion:              os.Getenv("AWS_REGION"),
			},
		},
	}

	// Call the handler
	return p.Handler(ctx, sqsEvent)
}

// convertMessageAttributes converts SQS message attributes
func convertMessageAttributes(attrs map[string]types.MessageAttributeValue) map[string]events.SQSMessageAttribute {
	result := make(map[string]events.SQSMessageAttribute)
	for k, v := range attrs {
		result[k] = events.SQSMessageAttribute{
			StringValue:      v.StringValue,
			BinaryValue:      v.BinaryValue,
			StringListValues: v.StringListValues,
			BinaryListValues: v.BinaryListValues,
			DataType:         aws.ToString(v.DataType),
		}
	}
	return result
}
