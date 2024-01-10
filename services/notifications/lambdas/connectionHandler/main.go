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
	db := dynamodb.New(session, aws.NewConfig().
		WithRegion("us-east-1").
		WithEndpoint("http://localstack:4566"),
	)

	return &DynamoClient{
		db: db,
	}
}

func (c *DynamoClient) PutConn(
	event events.APIGatewayWebsocketProxyRequest,
) events.APIGatewayV2HTTPResponse {
	tableName := os.Getenv("TABLE_NAME")

	// temporary hack to add userId to the connection
	// once rest of logic is confirmed, use cookies from session auth
	// and check that user has a session based on the cookies
	connectionUserId, ok := event.Headers["x-user-id"]
	if !ok {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Missing x-user-id header",
		}
	}

	fmt.Printf("Connection user id: %s\n", connectionUserId)
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
				"userId": {
					N: aws.String(connectionUserId),
				},
			},
		}

		_, err := c.db.PutItem(item)
		if err != nil {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       err.Error(),
			}
		}

		fmt.Printf("CONNECTED: %s with user %s\n", connId, connectionUserId)

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
				"userId": {
					N: aws.String(connectionUserId),
				},
			},
		}
		_, err := c.db.DeleteItem(item)
		if err != nil {
			fmt.Printf("\nUnable to delete connectionId from dynamo:\n %s\n", err.Error())
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       err.Error(),
			}
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
