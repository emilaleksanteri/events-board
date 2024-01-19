package main

import (
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

func (app *app) removePostLikeHandler(w http.ResponseWriter, r *http.Request) {
	tempUserId := int64(5)
	postId, err := app.getId(r, "id")
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err = app.models.Like.removePostLike(postId, tempUserId)
	if err != nil {
		switch err {
		case ErrRecordNotFound:
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(
		w,
		http.StatusOK,
		envelope{"message": "like removed successfully"},
		nil,
	)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) removeCommentLikeHandler(w http.ResponseWriter, r *http.Request) {
	tempUserId := int64(5)
	commentId, err := app.getId(r, "id")
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err = app.models.Like.removeCommentLike(commentId, tempUserId)
	if err != nil {
		switch err {
		case ErrRecordNotFound:
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(
		w,
		http.StatusOK,
		envelope{"message": "like removed successfully"},
		nil,
	)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
