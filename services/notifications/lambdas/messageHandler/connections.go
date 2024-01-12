package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type NotificationRow struct {
	ConnectionId string
	UserId       int64
}

func (app *App) getConnectionsForPost(senderUserId int64) (*[]NotificationRow, error) {
	friendIds, err := app.models.SocialConns.GetFriendsForUser(senderUserId)
	if err != nil {
		fmt.Printf("Could not get friends for user: %d\n", senderUserId)
		return nil, err
	}

	var filter expression.ConditionBuilder
	if len(friendIds) == 1 {
		filter = expression.Name("userId").Equal(expression.Value(friendIds[0]))
	} else {
		var expressionRows []expression.OperandBuilder
		for _, id := range friendIds {
			expressionRows = append(expressionRows, expression.Value(id))
		}

		filter = expression.Name("userId").
			In(expression.Value(expressionRows[0]), expressionRows...)
	}

	toGet := expression.NamesList(
		expression.Name("connectionId"),
		expression.Name("userId"),
	)

	expr, err := expression.NewBuilder().WithFilter(filter).WithProjection(toGet).Build()
	if err != nil {
		fmt.Printf("Got error building expression: %s\n", err.Error())
		return nil, err
	}

	allClients, err := app.dynamo.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(os.Getenv("TABLE_NAME")),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
	})

	if err != nil {
		fmt.Printf("Got error calling Scan: %s\n", err.Error())
		return nil, err
	}

	var rows []NotificationRow

	err = dynamodbattribute.UnmarshalListOfMaps(allClients.Items, &rows)
	if err != nil {
		fmt.Printf("Got error unmarshalling: %s\n", err.Error())
		return nil, err
	}

	return &rows, nil
}

func (app *App) getPostAuthorConnection(authorId int64) (*[]NotificationRow, error) {
	filter := expression.Name("userId").Equal(expression.Value(authorId))

	toGet := expression.NamesList(
		expression.Name("connectionId"),
		expression.Name("userId"),
	)

	expr, err := expression.NewBuilder().WithFilter(filter).WithProjection(toGet).Build()
	if err != nil {
		fmt.Printf("Got error building expression: %s\n", err.Error())
		return nil, err
	}

	// get many as a person might be connected on many tabs and there could be
	// old and new connections mixed
	possibleClients, err := app.dynamo.Scan(&dynamodb.ScanInput{
		TableName:                 aws.String(os.Getenv("TABLE_NAME")),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
	})

	if err != nil {
		fmt.Printf("Got error calling Scan: %s\n", err.Error())
		return nil, err
	}

	var rows []NotificationRow

	err = dynamodbattribute.UnmarshalListOfMaps(possibleClients.Items, &rows)
	if err != nil {
		fmt.Printf("Got error unmarshalling: %s\n", err.Error())
		return nil, err
	}

	return &rows, nil
}
