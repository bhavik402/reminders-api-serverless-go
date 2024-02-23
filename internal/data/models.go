package data

import "github.com/aws/aws-sdk-go-v2/service/dynamodb"

type DynamoDB struct {
	DyanmoClient *dynamodb.Client
}

type Models struct {
	Reminders ReminderModel
}

func NewModels(db *dynamodb.Client, tableName string) Models {
	return Models{
		Reminders: ReminderModel{DB: db, TableName: tableName},
	}
}
