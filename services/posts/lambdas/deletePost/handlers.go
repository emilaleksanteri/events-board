package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (app *app) deleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		app.notFoundHandler(w, r)
		return
	}

	err = app.models.Posts.delete(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(
		w,
		http.StatusOK,
		envelope{"message": "post successfully deleted"},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
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
