package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"events/posts/models"
	"github.com/aws/aws-lambda-go/events"
	"github.com/go-chi/chi/v5"
)

func Handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return chiLambda.ProxyWithContext(ctx, event)
}

func (app *app) listPostsHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	take, err := app.readInt(qs, "take", 10)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	skip, err := app.readInt(qs, "skip", 0)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	posts, metadata, err := app.models.Posts.List(take, skip)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"posts": posts, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) getPostHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		app.notFoundHandler(w, r)
		return
	}

	qs := r.URL.Query()
	take, err := app.readInt(qs, "take", 10)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	skip, err := app.readInt(qs, "skip", 0)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	post, err := app.models.Posts.Get(int64(id), take, skip)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"post": post}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

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
