package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type DynamoClient struct {
	db *dynamodb.DynamoDB
}

func NewDymanoDbClient() *DynamoClient {
	session := session.Must(session.NewSession())
	db := dynamodb.New(session)

	return &DynamoClient{
		db: db,
	}
}

func (c *DynamoClient) PutConn(
	event events.APIGatewayWebsocketProxyRequest,
) events.APIGatewayProxyResponse {
	tableName := os.Getenv("TABLE_NAME")
	eventType := event.RequestContext.EventType
	connId := event.RequestContext.ConnectionID

	switch eventType {
	case "CONNECT":
		oneHourFromNow := time.Now().Add(1 * time.Hour)
		item := &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]*dynamodb.AttributeValue{
				"connectionId": {
					S: aws.String(connId),
				},
				"ttl": {
					N: aws.String(fmt.Sprintf("%d", oneHourFromNow.Unix())),
				},
				"notificationId": {
					S: aws.String("DEFAULT"),
				},
			},
		}

		c.db.PutItem(item)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       "Connected.",
		}

	case "DISCONNECT":
		item := &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"connectionId": {
					S: aws.String(connId),
				},
			},
		}
		c.db.DeleteItem(item)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       "Disconnected.",
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "Ok.",
	}
}

type App struct {
	db *DynamoClient
}

func (app *App) handler(
	event events.APIGatewayWebsocketProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return app.db.PutConn(event), nil
}

func main() {
	dbClient := NewDymanoDbClient()
	app := &App{
		db: dbClient,
	}
	lambda.Start(app.handler)
}
