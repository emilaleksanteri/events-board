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
	db := dynamodb.New(session, aws.NewConfig().WithDisableSSL(true))

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

		output, err := c.db.PutItem(item)
		if err != nil {
			fmt.Printf("Error: %s", err.Error())
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       err.Error(),
			}
		}

		fmt.Printf("Save conn %s:\n %+v", connId, output)
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
			},
		}
		c.db.DeleteItem(item)
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
