package main

import (
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

func (app *app) getPostLikes(w http.ResponseWriter, r *http.Request) {
	postId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || postId < 1 {
		app.errorResponse(w, r, http.StatusBadRequest, errors.New("invalid post id!"))
		return
	}

	filter := &Filter{}
	qs := r.URL.Query()
	filter.take, err = strconv.Atoi(qs.Get("take"))
	if err != nil {
		filter.take = 10
	}

	filter.skip, err = strconv.Atoi(qs.Get("skip"))
	if err != nil {
		filter.skip = 0
	}

	likes, err := app.models.Like.getLikes(postId, filter)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"post_likes": likes}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
