package main

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]any

func (app *app) writeJSON(
	w http.ResponseWriter,
	status int,
	data envelope,
	headers http.Header,
) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *app) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		w.WriteHeader(500)
	}
}

func (app *app) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	message := "the server encountared a problem and could not process this request :("
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}
