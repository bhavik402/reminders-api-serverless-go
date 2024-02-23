package app

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bhavik3210/reminders/serverless-reminders-rest-api/internal/data"
	"github.com/google/uuid"
)

const (
	RES_REMINDERS       = "/api/v1/reminders"
	RES_A_REMINDER      = "/api/v1/reminders/{remindersId}"
	RES_REMINDER_STATUS = "/api/v1/reminders/status/{remindersId}"
	RES_REMINDER_FLAG   = "/api/v1/reminders/flag/{remindersId}"
)

func (app Application) HandleRoutes(event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch event.HTTPMethod {
	case "GET":
		switch event.Resource {
		case RES_REMINDERS:
			return app.getReminders(event)
		case RES_A_REMINDER:
			return app.getAReminder(event)
		}
	case "POST":
		return app.postReminder(event)
	case "PUT":
		switch event.Resource {
		case RES_REMINDER_STATUS:
			return app.updateReminderStatus(event)
		case RES_REMINDER_FLAG:
			return app.updateReminderFlag(event)
		}
	case "DELETE":
		return app.deleteReminder(event)
	}
	return NotSupported("Not Supported"), nil
}

func (app Application) getReminders(event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	res, err := app.Models.Reminders.ReadAll()
	if err != nil {
		return InternalServerError(fmt.Sprintf("failed to query all reminders: %s", err.Error())), nil
	}

	response, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return InternalServerError(fmt.Sprintf("failed to Marshal: %s", err.Error())), nil
	}

	return Ok(string(response)), nil
}

func (app Application) getAReminder(event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	res, err := app.Models.Reminders.ReadAReminderById(event.PathParameters["remindersId"])
	if err != nil {
		return InternalServerError(fmt.Sprintf("failed to retrieve a remidner: %s", err.Error())), nil
	}

	response, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return InternalServerError(fmt.Sprintf("failed to Marshal: %s", err.Error())), nil
	}

	return Ok(string(response)), nil
}

func (app Application) postReminder(event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	r := data.Reminder{}
	json.Unmarshal([]byte(event.Body), &r)
	r.ID = uuid.New().String()
	err := app.Models.Reminders.Insert(r)
	if err != nil {
		return InternalServerError(fmt.Sprintf("failed to create new task: %s", err.Error())), nil
	}
	return Ok("Reminder Insertion Success"), nil
}

func (app Application) deleteReminder(event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return Ok("deleteReminder"), nil
}

func (app Application) updateReminderStatus(event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return Ok("updateReminderStatus"), nil
}

func (app Application) updateReminderFlag(event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return Ok("updateReminderFlag"), nil
}
