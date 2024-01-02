package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	chiLambda         *chiadapter.ChiLambda
)

func (app *app) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	err := app.writeJSON(w, http.StatusOK, envelope{"status": "available"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *app) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, "resource not found")
}

func (app *app) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	tempUserId := int64(2)
	var input struct {
		Body   string `json:"body"`
		PostId int64  `json:"post_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Body == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "body must not be blank")
		return
	}

	if input.PostId < 1 {
		app.errorResponse(w, r, http.StatusBadRequest, "post_id must be a valid integer")
		return
	}

	comment := &Comment{
		Body:   input.Body,
		PostId: input.PostId,
	}

	err = app.models.Comments.insertRootComment(comment, tempUserId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/comments/%d", comment.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) createSubCommentHandler(w http.ResponseWriter, r *http.Request) {
	parentId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid comment id")
		return
	}

	tempUserId := int64(2)
	var input struct {
		Body   string `json:"body"`
		PostId int64  `json:"post_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Body == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "body must not be blank")
		return
	}

	if input.PostId < 1 {
		app.errorResponse(w, r, http.StatusBadRequest, "post_id must be a valid integer")
		return
	}

	comment := &Comment{
		Body:   input.Body,
		PostId: input.PostId,
	}

	err = app.models.Comments.insertSubComment(comment, tempUserId, int64(parentId))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/comments/%d", comment.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

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
	r.Route("/create", func(r chi.Router) {
		r.Post("/", app.createCommentHandler)
		r.Post("/{id}", app.createSubCommentHandler)

		r.Get("/healthcheck", app.healthcheckHandler)
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
