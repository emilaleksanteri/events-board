package main

import (
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
