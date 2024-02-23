package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bhavik3210/reminders/serverless-reminders-rest-api/pkg/app"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	app := app.NewApplication(event.RequestContext.Stage)
	return app.HandleRoutes(event)
}
