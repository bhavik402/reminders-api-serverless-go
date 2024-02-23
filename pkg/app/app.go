package app

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/bhavik3210/reminders/serverless-reminders-rest-api/internal/data"
	"github.com/charmbracelet/log"
)

type AppConfig struct {
	Env       string
	TableName string
	AWSRegion string
}

type Application struct {
	Config AppConfig
	Models data.Models
}

func NewApplication(stage string) Application {
	cfg := AppConfig{
		Env:       stage,
		TableName: "Reminders",
		AWSRegion: "us-east-1",
	}

	dbCfg, err := config.LoadDefaultConfig(
		context.Background(),
		func(opts *config.LoadOptions) error {
			opts.Region = cfg.AWSRegion
			return nil
		},
	)
	if err != nil {
		log.Error("Failed to Load DynamoDB Config %w", err)
	}

	m := data.NewModels(dynamodb.NewFromConfig(dbCfg), cfg.TableName)

	return Application{Config: cfg, Models: m}
}

func (app *Application) ResourceResolver(event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {

	b, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return nil, err
	}

	// _ = helperFunc(event)

	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(b),
	}, nil
}
