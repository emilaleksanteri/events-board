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

func (app *app) follow(w http.ResponseWriter, r *http.Request) {
	tempUserId := int64(4)
	toFollow := chi.URLParam(r, "user")

	userId, err := strconv.ParseInt(toFollow, 10, 64)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid user ID")
		return
	}

	following, err := app.models.User.GetUser(userId)
	if err != nil {
		switch err {
		case ErrRecordNotFound:
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	follower := User{Id: tempUserId}

	err = app.models.Social.Follow(&follower, &following)
	if err != nil {
		switch {
		case errors.Is(err, ErrAlreadyFollowing):
			app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// TODO add header to users followers endpoint

	err = app.writeJSON(w, http.StatusOK, envelope{"status": "followed"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
