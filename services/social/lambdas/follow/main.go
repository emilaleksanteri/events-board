package main

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

var chiLambda *chiadapter.ChiLambda

type app struct {
	models Models
	eb     *eventbridge.EventBridge
}

func NewEventBridge() *eventbridge.EventBridge {
	session := session.Must(session.NewSession())
	eb := eventbridge.New(session, aws.NewConfig().
		WithRegion("us-east-1").
		WithEndpoint("http://localstack:4566"),
	)

	return eb
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

func init() {
	db, err := openDB()
	if err != nil {
		panic(err)
	}

	app := app{models: NewModels(db), eb: NewEventBridge()}
	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Post("/follow/{user}", app.follow)
	})

	r.NotFound(app.notFoundHandler)

	chiLambda = chiadapter.New(r)
}

func Handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return chiLambda.ProxyWithContext(ctx, event)
}

func main() {
	lambda.StartWithOptions(Handler, lambda.WithContext(context.Background()))
}
