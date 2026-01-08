package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jobping/backend/internal/app"
	userfanouthandler "github.com/jobping/backend/internal/features/user_fanout/handler"
)

var sqsHandler *userfanouthandler.SQSHandler

func init() {
	appInstance, err := app.BuildUserFanout()
	if err != nil {
		log.Fatalf("Failed to build application: %v", err)
	}

	sqsHandler = appInstance.SQSHandler
}

func handler(ctx context.Context, event json.RawMessage) (interface{}, error) {
	var sqsEvent events.SQSEvent
	if err := json.Unmarshal(event, &sqsEvent); err != nil {
		log.Printf("Failed to parse SQS event: %v", err)
		return nil, err
	}

	log.Printf("Processing SQS event with %d records", len(sqsEvent.Records))
	return nil, sqsHandler.HandleSQSEvent(ctx, sqsEvent)
}

func main() {
	lambda.Start(handler)
}


