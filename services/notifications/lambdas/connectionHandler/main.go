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
	// TODO separate dev and prod configs
	db := dynamodb.New(session, aws.NewConfig().WithRegion("us-east-1").WithEndpoint("http://localstack:4566"))

	return &DynamoClient{
		db: db,
	}
}

func (c *DynamoClient) PutConn(
	event events.APIGatewayWebsocketProxyRequest,
) events.APIGatewayV2HTTPResponse {
	tableName := os.Getenv("TABLE_NAME")
	connId := event.RequestContext.ConnectionID
	eventType := event.RequestContext.EventType
	if eventType == "CONNECT" {
		fmt.Printf("table name: %s\n", tableName)
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

		_, err := c.db.PutItem(item)
		if err != nil {
			fmt.Printf("Error bruh moment: %s\n\n", err.Error())
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       err.Error(),
			}
		}

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       connId,
		}
	}

	if eventType == "DISCONNECT" {
		item := &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"connectionId": {
					S: aws.String(connId),
				},
				"notificationId": {
					S: aws.String("DEFAULT"),
				},
			},
		}
		_, err := c.db.DeleteItem(item)
		if err != nil {
			fmt.Printf("\nUnable to delete connectionId from dynamo:\n %s\n", err.Error())
		}

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       "Disconnected.",
		}
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       "Ok.",
	}
}

type App struct {
	db *DynamoClient
}

func (app *App) handler(
	event events.APIGatewayWebsocketProxyRequest,
) (events.APIGatewayV2HTTPResponse, error) {
	return app.db.PutConn(event), nil
}

func main() {
	dbClient := NewDymanoDbClient()
	app := &App{
		db: dbClient,
	}
	lambda.Start(app.handler)
}
