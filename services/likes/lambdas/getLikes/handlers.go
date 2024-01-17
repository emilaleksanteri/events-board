package main

import (
	"errors"
	"fmt"
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

// TODO, these handlers could be probably be refactored into one

func (app *app) postLikesHandler(w http.ResponseWriter, r *http.Request) {
	postId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || postId < 1 {
		app.errorResponse(w, r, http.StatusBadRequest, errors.New("invalid post id!"))
		return
	}

	qs := r.URL.Query()
	filter := &Filter{
		take: app.getIntParam(qs, "take", 30),
		skip: app.getIntParam(qs, "skip", 0),
	}

	likes, err := app.models.Like.getPostLikes(postId, filter)
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
	commentId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || commentId < 1 {
		app.errorResponse(w, r, http.StatusBadRequest, errors.New("invalid comment id!"))
		return
	}

	qs := r.URL.Query()
	filter := &Filter{
		take: app.getIntParam(qs, "take", 30),
		skip: app.getIntParam(qs, "skip", 0),
	}

	likes, err := app.models.Like.getCommentLikes(commentId, filter)
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
