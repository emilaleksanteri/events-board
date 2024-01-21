package main

import (
	"errors"
	"net/http"
	"strconv"

	"getComment/models"
	"github.com/go-chi/chi/v5"
)

func (app *app) getCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid comment id parameter"))
		return
	}
	qs := r.URL.Query()
	take, err := app.readInt(qs, "take", 10)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	offset, err := app.readInt(qs, "offset", 0)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	comment, err := app.models.Comments.GetComment(commentId, take, offset)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comment": comment}, nil)
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
