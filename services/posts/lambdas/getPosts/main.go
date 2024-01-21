package main

import (
	"context"
	"net/http"
	"os"

	"events/posts/models"
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

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		app.writeJSON(w, http.StatusOK, envelope{"message": "hello from root"}, nil)
	})

	r.Route("/posts", func(r chi.Router) {
		r.Get("/", app.listPostsHandler)
		r.Get("/{id}", app.getPostHandler)
		r.Get("/healthcheck", app.healthcheckHandler)
	})

	r.NotFound(app.notFoundHandler)
	chiLambda = chiadapter.New(r)
}

func main() {
	lambda.StartWithOptions(Handler, lambda.WithContext(context.Background()))
}
