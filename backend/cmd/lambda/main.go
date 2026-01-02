package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
	"github.com/jobping/backend/internal/app"
	jobhandler "github.com/jobping/backend/internal/features/job/handler"
)

var (
	chiLambda  *chiadapter.ChiLambda
	sqsHandler *jobhandler.SQSHandler
)

func init() {
	appInstance, err := app.Build()
	if err != nil {
		log.Fatalf("Failed to build application: %v", err)
	}

	router, ok := appInstance.Router.(*chi.Mux)
	if !ok {
		log.Fatalf("Router is not a chi.Mux")
	}

	chiLambda = chiadapter.New(router)
	sqsHandler = appInstance.SQSHandler
}

// handler routes Lambda events to either HTTP or SQS handler
func handler(ctx context.Context, event json.RawMessage) (interface{}, error) {
	// Try to detect event type
	var sqsEvent events.SQSEvent
	if err := json.Unmarshal(event, &sqsEvent); err == nil && len(sqsEvent.Records) > 0 {
		// This is an SQS event
		log.Printf("Processing SQS event with %d records", len(sqsEvent.Records))
		return nil, sqsHandler.HandleSQSEvent(ctx, sqsEvent)
	}

	// Otherwise, treat as API Gateway event
	var apiEvent events.APIGatewayProxyRequest
	if err := json.Unmarshal(event, &apiEvent); err != nil {
		log.Printf("Failed to parse event: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	return chiLambda.ProxyWithContext(ctx, apiEvent)
}

func main() {
	lambda.Start(handler)
}
