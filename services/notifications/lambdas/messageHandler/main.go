package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	_ "github.com/lib/pq"
)

func NewDymanoDbClient() *dynamodb.DynamoDB {
	session := session.Must(session.NewSession())
	db := dynamodb.New(session, aws.NewConfig().
		WithRegion("us-east-1").
		WithEndpoint("http://localstack:4566"),
	)

	return db
}

func NewGatewayClient() *apigatewaymanagementapi.ApiGatewayManagementApi {
	session := session.Must(session.NewSession())
	gw := apigatewaymanagementapi.New(session, aws.NewConfig().
		WithRegion("us-east-1").
		WithEndpoint("http://localstack:4566"),
	)

	return gw
}

func openDB() (*sql.DB, error) {
	addr := os.Getenv("DB_ADDRESS")
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

type App struct {
	dynamo *dynamodb.DynamoDB
	gw     *apigatewaymanagementapi.ApiGatewayManagementApi
	models Models
}

func main() {
	dbClient := NewDymanoDbClient()
	gwClient := NewGatewayClient()
	pgDb, err := openDB()
	if err != nil {
		fmt.Printf("Could not open db: %s\n", err.Error())
		return
	}

	app := &App{
		dynamo: dbClient,
		gw:     gwClient,
		models: NewModels(pgDb),
	}

	lambda.Start(app.handler)
}
