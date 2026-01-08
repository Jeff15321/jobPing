package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
	"github.com/jobping/backend/internal/app"
)

var chiLambda *chiadapter.ChiLambda

func init() {
	appInstance, err := app.BuildAPI()
	if err != nil {
		log.Fatalf("Failed to build application: %v", err)
	}

	router, ok := appInstance.Router.(*chi.Mux)
	if !ok {
		log.Fatalf("Router is not a chi.Mux")
	}

	chiLambda = chiadapter.New(router)
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return chiLambda.ProxyWithContext(ctx, event)
}

func main() {
	lambda.Start(handler)
}


