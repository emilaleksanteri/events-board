package main

import (
	"encoding/json"
	//"errors"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	//"github.com/aws/aws-sdk-go/service/eventbridge"
)

type DynamoClient struct {
	db *dynamodb.DynamoDB
}

func NewDymanoDbClient() *DynamoClient {
	session := session.Must(session.NewSession())
	db := dynamodb.New(session, aws.NewConfig().WithRegion("us-east-1").WithEndpoint("http://localstack:4566"))

	return &DynamoClient{
		db: db,
	}
}

type GatewayClient struct {
	gw *apigatewaymanagementapi.ApiGatewayManagementApi
}

func NewGatewayClient() *GatewayClient {
	session := session.Must(session.NewSession())
	//endPoint := os.Getenv("ENDPOINT")
	gw := apigatewaymanagementapi.New(session, aws.NewConfig().WithRegion("us-east-1").WithEndpoint("http://localstack:4566"))

	return &GatewayClient{
		gw: gw,
	}
}

type App struct {
	db *DynamoClient
	gw *GatewayClient
}

type NotificationRow struct {
	NotificationId string
	ConnectionId   string
}

func (app *App) getConnections(senderConnId, notificationId string) (*[]NotificationRow, error) {
	tableName := os.Getenv("TABLE_NAME")
	filter := expression.Name("notificationId").Equal(expression.Value(notificationId))

	toGet := expression.NamesList(
		expression.Name("connectionId"),
		expression.Name("notificationId"),
	)

	expr, err := expression.NewBuilder().WithFilter(filter).WithProjection(toGet).Build()
	if err != nil {
		fmt.Printf("\nGot error building expression: %s\n", err.Error())
		return nil, err
	}

	allClients, err := app.db.db.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(tableName),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
	})

	if err != nil {
		fmt.Printf("\nGot error calling Scan: %s\n", err.Error())
		return nil, err
	}

	var rows []NotificationRow

	err = dynamodbattribute.UnmarshalListOfMaps(allClients.Items, &rows)
	if err != nil {
		fmt.Printf("\nGot error unmarshalling: %s\n", err.Error())
		return nil, err
	}

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
	if err != nil {
		fmt.Printf("\nCould not unmarshal event: %v\n", event.Detail)
		return err
	}

	conns, err := app.getConnections(eventData.ConnectionId, eventData.NotificationId)
	if err != nil {
		fmt.Printf("\nCould not get connections: %v\n", eventData)
		return err
	}
	for _, conn := range *conns {
		//go func(conn NotificationRow) {
		//dataToSend, err := json.Marshal(eventData)
		//if err != nil {
		//	return err
		//}
		fmt.Printf("sending data: %+v\n connId: %s\n", conn, conn.ConnectionId)

		_, err = app.gw.gw.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(conn.ConnectionId),
			Data:         event.Detail,
		})
		if err != nil {
			fmt.Printf("\nCould not send a msg to a conn: %s\n", err.Error())
		}
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
