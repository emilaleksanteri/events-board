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

func (app *app) postLikesHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.getId(r, "id")
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	likes, err := app.models.Like.getPostLikes(id, app.getFilter(r))
	if err != nil {
		fmt.Printf("failed to get post likes: %s\n", err.Error())
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"post_likes": likes}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) commentLikesHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.getId(r, "id")
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	likes, err := app.models.Like.getCommentLikes(id, app.getFilter(r))
	if err != nil {
		fmt.Printf("failed to get comment likes: %s\n", err.Error())
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comment_likes": likes}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
