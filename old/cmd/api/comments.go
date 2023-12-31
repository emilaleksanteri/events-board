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

	err = app.models.Comments.Insert(comment, 2)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.publishPostCommentEvent(comment, r.Context())
	if err != nil {
		app.logger.Error(err.Error())
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

	err = app.models.Comments.InsertSubComment(comment, id, 2)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.publishPostCommentEvent(comment, r.Context())
	if err != nil {
		app.logger.Error(err.Error())
	}

	err = app.publishPostSubCommentEvent(comment, r.Context())
	if err != nil {
		app.logger.Error(err.Error())
	}
}

func (app *application) getCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	var input data.Filters

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	valid := validator.New()
	qs := r.URL.Query()

	input.Take = app.readInt(qs, "take", 10, valid)
	input.Offset = app.readInt(qs, "offset", 0, valid)

	if data.ValidateFileters(valid, input); !valid.Valid() {
		app.failedValidationResponse(w, r, valid.Errors)
		return
	}

	comment, err := app.models.Comments.Get(id, &input)
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

func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Comments.DeleteComment(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "comment deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) handleSubscribeToComments(w http.ResponseWriter, r *http.Request) {
	err := app.handleServerEvents(w, r, COMMENT_ADDED)
	if err != nil {
		switch {
		case errors.Is(err, ErrSseNotSupported):
			app.sSENotSupportedResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}
}
