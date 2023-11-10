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
	router.HandlerFunc(http.MethodGet, "/v1/posts/:id", app.getPostHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/posts/:id", app.deletePostHandler)

	router.HandlerFunc(http.MethodPost, "/v1/comments", app.createCommentHandler)
	router.HandlerFunc(http.MethodPost, "/v1/comments/:id", app.createSubCommentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/comments/:id", app.getCommentHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/comments/:id", app.deleteCommentHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(router)))
}
