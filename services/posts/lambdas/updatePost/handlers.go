package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

func (app *app) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		app.notFoundHandler(w, r)
		return
	}

	var input struct {
		Body string `json:"body"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if input.Body == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid body")
		return
	}

	if len(input.Body) > 20_000 {
		app.errorResponse(
			w,
			r,
			http.StatusBadRequest,
			"body too long, max 20_000 characters",
		)
		return
	}

	post, err := app.models.Posts.Get(int64(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	post.Body = input.Body
	err = app.models.Posts.Update(post)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundHandler(w, r)
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

