package main

import (
	"errors"
	"net/http"

	"github.com/emilaleksanteri/pubsub/internal/data"
	"github.com/emilaleksanteri/pubsub/internal/validator"
)

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Body   string `json:"body"`
		PostId int64  `json:"post_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	valid := validator.New()
	comment := &data.Comment{
		Body:   input.Body,
		PostId: input.PostId,
	}

	data.ValidateComment(valid, comment)
	if !valid.Valid() {
		app.failedValidationResponse(w, r, valid.Errors)
		return
	}

	err = app.models.Comments.Insert(comment)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createSubCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var input struct {
		Body   string `json:"body"`
		PostId int64  `json:"post_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	valid := validator.New()

	comment := &data.Comment{
		Body:   input.Body,
		PostId: input.PostId,
	}

	data.ValidateComment(valid, comment)
	if !valid.Valid() {
		app.failedValidationResponse(w, r, valid.Errors)
		return
	}

	err = app.models.Comments.InsertSubComment(comment, id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) getCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	comment, err := app.models.Comments.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
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
