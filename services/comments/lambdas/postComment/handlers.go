package main

import (
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

func (app *app) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	tempUserId := int64(4)
	var input struct {
		Body   string `json:"body"`
		PostId int64  `json:"post_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Body == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "body must not be blank")
		return
	}

	if input.PostId < 1 {
		app.errorResponse(w, r, http.StatusBadRequest, "post_id must be a valid integer")
		return
	}

	comment := &Comment{
		Body:   input.Body,
		PostId: input.PostId,
	}

	err = app.models.Comments.insertRootComment(comment, tempUserId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	go func(comment *Comment) {
		err = app.publishComment(comment)
		if err != nil {
			fmt.Printf("Error publishing comment event: %s\n", err.Error())
		}
	}(comment)

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/comments/%d", comment.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *app) createSubCommentHandler(w http.ResponseWriter, r *http.Request) {
	parentId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid comment id")
		return
	}

	tempUserId := int64(5)
	var input struct {
		Body   string `json:"body"`
		PostId int64  `json:"post_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Body == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "body must not be blank")
		return
	}

	if input.PostId < 1 {
		app.errorResponse(w, r, http.StatusBadRequest, "post_id must be a valid integer")
		return
	}

	comment := &Comment{
		Body:   input.Body,
		PostId: input.PostId,
	}

	err = app.models.Comments.insertSubComment(comment, tempUserId, int64(parentId))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	go func(comment *Comment) {
		err := app.publishChildComment(comment)
		if err != nil {
			fmt.Printf("Error publishing comment event: %s\n", err.Error())
		}
	}(comment)

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/comments/%d", comment.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
