package main

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	//"github.com/aws/aws-sdk-go/service/eventbridge"
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

type GatewayClient struct {
	gw *apigatewaymanagementapi.ApiGatewayManagementApi
}

func NewGatewayClient() *GatewayClient {
	session := session.Must(session.NewSession())
	gw := apigatewaymanagementapi.New(session)

	return &GatewayClient{
		gw: gw,
	}
}

type App struct {
	db *DynamoClient
	gw *GatewayClient
}

type NotificationRow struct {
	notificationId string
	connectionId   string
	ttl            int64
}

func (app *App) getConnections(senderConnId, notificationId string) (*[]NotificationRow, error) {
	tableName := os.Getenv("TABLE_NAME")
	allClients, err := app.db.db.Scan(&dynamodb.ScanInput{
		TableName: aws.String(tableName),
		FilterExpression: aws.String(
			"connectionId <> :connectionId AND notificationId = :notificationId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":connectionId": {
				S: aws.String(senderConnId),
			},
			":notificationId": {
				S: aws.String(notificationId),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	rows := []NotificationRow{}
	err = dynamodbattribute.UnmarshalListOfMaps(allClients.Items, &rows)
	return &rows, nil
}

type EventBridgeEvent struct {
	ConnectionId   string `json:"connectionId"`
	NotificationId string `json:"notificationId"`
	Message        string `json:"message"`
}

func (app *App) handler(
	event events.CloudWatchEvent,
) error {
	var eventData EventBridgeEvent
	err := json.Unmarshal(event.Detail, &eventData)

	connId := "1" //event.RequestContext.ConnectionID
	conns, err := app.getConnections(connId, "DEFAULT")
	if err != nil {
		return err
	}

	for _, conn := range *conns {
		//go func(conn NotificationRow) {
		dataToSend, err := json.Marshal(eventData)
		if err != nil {
			return err
		}

		app.gw.gw.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(conn.connectionId),
			Data:         dataToSend,
		})
		//}(conn)
	}

	return nil
}

func main() {
	dbClient := NewDymanoDbClient()
	gwClient := NewGatewayClient()
	app := &App{
		db: dbClient,
		gw: gwClient,
	}
	lambda.Start(app.handler)
}
