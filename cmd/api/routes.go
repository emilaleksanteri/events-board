package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowdResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodPost, "/v1/posts", app.cratePostHandler)
	router.HandlerFunc(http.MethodGet, "/v1/posts", app.listPostsHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(router)))
}
