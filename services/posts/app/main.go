package main

import (
	"context"
	"os"

	"database/sql"
	"time"

	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

var chiLambda *chiadapter.ChiLambda

type app struct {
	models Models
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

	app := app{models: NewModels(db)}
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		app.writeJSON(w, http.StatusOK, envelope{"message": "hello from lambda hot reload"}, nil)
	})

	r.Get("/healthcheck", app.healthcheckHandler)

	r.Route("/posts", func(r chi.Router) {
		r.Get("/", app.listPostsHandler)
		r.Get("/{id}", app.getPostHandler)
	})

	r.NotFound(app.notFoundHandler)
	chiLambda = chiadapter.New(r)
}

func main() {
	lambda.StartWithOptions(Handler, lambda.WithContext(context.Background()))
}
