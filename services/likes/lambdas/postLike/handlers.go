package main

import (
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

func (app *app) likePostHandler(w http.ResponseWriter, r *http.Request) {
	tempUserId := int64(5)
	postId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	postLike := &PostLike{
		PostId: postId,
		UserId: tempUserId,
	}

	err = app.models.Like.likePost(postLike)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"post_like": postLike}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *app) likeCommentHandler(w http.ResponseWriter, r *http.Request) {
	tempUserId := int64(5)
	commentId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	commentLike := &CommentLike{
		CommentId: commentId,
		UserId:    tempUserId,
	}

	err = app.models.Like.likeComment(commentLike)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment_like": commentLike}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
