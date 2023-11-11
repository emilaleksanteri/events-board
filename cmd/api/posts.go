package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/emilaleksanteri/pubsub/internal/data"
	"github.com/emilaleksanteri/pubsub/internal/validator"
)

func (app *application) cratePostHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Body string `json:"body"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	valid := validator.New()

	post := &data.Post{
		Body: input.Body,
	}

	data.ValidPost(valid, post)

	if !valid.Valid() {
		app.failedValidationResponse(w, r, valid.Errors)
		return
	}

	err := app.models.Posts.Insert(post)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/posts/%d", post.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"post": post}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.publishPostPostEvent(post, r.Context())
	if err != nil {
		app.logger.Info("failed to publish post event", "error", err)
	}
}

func (app *application) listPostsHandler(w http.ResponseWriter, r *http.Request) {
	var input data.Filters

	v := validator.New()
	qs := r.URL.Query()

	input.Take = app.readInt(qs, "take", 20, v)
	input.Offset = app.readInt(qs, "offset", 0, v)

	if data.ValidateFileters(v, input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	posts, metadata, err := app.models.Posts.GetAll(input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"posts": posts, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	var input data.Filters

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	valid := validator.New()
	qs := r.URL.Query()

	input.Take = app.readInt(qs, "take", 10, valid)
	input.Offset = app.readInt(qs, "offset", 0, valid)

	if data.ValidateFileters(valid, input); !valid.Valid() {
		app.failedValidationResponse(w, r, valid.Errors)
		return
	}

	post, err := app.models.Posts.Get(id, &input)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"post": post}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Posts.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "post deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) handleSubscribeToPosts(w http.ResponseWriter, r *http.Request) {
	err := app.handleServerEvents(w, r, POST_ADDED)
	if err != nil {
		switch {
		case errors.Is(err, ErrSseNotSupported):
			app.sSENotSupportedResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}
}
