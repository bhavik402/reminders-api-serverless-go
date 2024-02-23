package data

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Reminder struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type ReminderDB struct {
	ID    string `dynamodbav:"pk"`
	Title string `dynamodbav:"sk"`
}

type ReminderModel struct {
	DB        *dynamodb.Client
	TableName string
}

func (rm ReminderModel) Insert(reminder Reminder) error {
	rdbm := reminder.toDBModel()

	item, err := attributevalue.MarshalMap(rdbm)
	if err != nil {
		return fmt.Errorf("failed to Marshalmap in DynamoDB: %w", err)
	}

	_, err = rm.DB.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(rm.TableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to PutItem in DyanmoDB: %w", err)
	}

	return nil
}

func (rm ReminderModel) ReadAll() ([]Reminder, error) {
	var reminders []Reminder
	var response *dynamodb.ScanOutput
	// filtEx := expression.Name("year").Between(expression.Value(startYear), expression.Value(endYear))
	projEx := expression.NamesList(
		expression.Name("pk"), expression.Name("sk"))
	expr, err := expression.NewBuilder().WithProjection(projEx).Build()
	if err != nil {
		log.Printf("Couldn't build expressions for scan. Here's why: %v\n", err)
	} else {
		scanPaginator := dynamodb.NewScanPaginator(rm.DB, &dynamodb.ScanInput{
			TableName:                 aws.String(rm.TableName),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ProjectionExpression:      expr.Projection(),
		})
		for scanPaginator.HasMorePages() {
			response, err = scanPaginator.NextPage(context.TODO())
			if err != nil {
				// log.Printf("Couldn't scan for movies released between %v and %v. Here's why: %v\n",
				// 	pk, sk, err)
				// break
				return nil, err
			} else {
				var reminderPage []ReminderDB
				err = attributevalue.UnmarshalListOfMaps(response.Items, &reminderPage)
				if err != nil {
					// log.Printf("Couldn't unmarshal query response. Here's why: %v\n", err)
					// break
					return nil, err
				} else {
					for _, r := range reminderPage {
						reminders = append(reminders, *r.fromDBModel())
					}
				}
			}
		}
	}
	return reminders, err
}

func (rm ReminderModel) ReadAReminderById(id string) ([]Reminder, error) {
	var reminders []Reminder

	keyEx := expression.Key("pk").Equal(expression.Value(id))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build dynamodb expression %w", err)
	}

	queryPaginator := dynamodb.NewQueryPaginator(rm.DB, &dynamodb.QueryInput{
		TableName:                 &rm.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	for queryPaginator.HasMorePages() {
		res, err := queryPaginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("couldn't query: %w", err)
		}

		var reminderPage []ReminderDB
		err = attributevalue.UnmarshalListOfMaps(res.Items, &reminderPage)
		if err != nil {
			log.Printf("Couldn't unmarshal query response. Here's why: %v\n", err)
			break
		} else {
			for _, r := range reminderPage {
				reminders = append(reminders, *r.fromDBModel())
			}

		}
	}

	return reminders, err
}

func (r Reminder) toDBModel() *ReminderDB {
	return &ReminderDB{
		ID:    r.ID,
		Title: r.Title,
	}
}

func (r ReminderDB) fromDBModel() *Reminder {
	return &Reminder{
		ID:    r.ID,
		Title: r.Title,
	}
}
