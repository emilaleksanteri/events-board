package main

import (
	"context"
	"os"

	"getComment/models"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

var chiLambda *chiadapter.ChiLambda

type app struct {
	models models.Models
}

func init() {
	addr := os.Getenv("DB_ADDRESS")
	db, err := models.OpenDB(addr)
	if err != nil {
		panic(err)
	}

	app := app{models: models.NewModels(db)}
	r := chi.NewRouter()
	r.Route("/comments", func(r chi.Router) {
		r.Get("/healthcheck", app.healthcheckHandler)
		r.Get("/{id}", app.getCommentHandler)
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
