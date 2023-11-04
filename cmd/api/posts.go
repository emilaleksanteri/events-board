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
