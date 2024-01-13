package main

import (
	"fmt"
	"net/http"
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

func (app *app) createHandler(w http.ResponseWriter, r *http.Request) {
	tempUsrId := int64(3)
	var input struct {
		Body string `json:"body"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if input.Body == "" {
		app.errorResponse(
			w,
			r,
			http.StatusBadRequest,
			"missing body, min length is 1 character",
		)
		return
	}

	if len(input.Body) > 20_000 {
		app.errorResponse(
			w,
			r,
			http.StatusBadRequest,
			"body too long, max is 20_000 characters",
		)
		return
	}

	post := &Post{
		Body: input.Body,
	}

	err = app.models.Posts.Insert(post, tempUsrId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	go func(post *Post) {
		err = app.publishPost(post)
		if err != nil {
			fmt.Printf("Could not publish event for post %d: \n%v\n", post.Id, err)
		}
	}(post)

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/posts/%d", post.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"post": post}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
