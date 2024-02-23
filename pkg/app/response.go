package app

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func Ok(body string) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       body,
	}
}

func NotSupported(body string) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotFound,
		Body:       body,
	}
}

func InternalServerError(body string) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       body,
	}
}
