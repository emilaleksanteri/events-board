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
	router.HandlerFunc(http.MethodGet, "/v1/subscribe/posts", app.handleSubscribeToPosts)

	router.HandlerFunc(http.MethodPost, "/v1/comments", app.createCommentHandler)
	router.HandlerFunc(http.MethodPost, "/v1/comments/:id", app.createSubCommentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/comments/:id", app.getCommentHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/comments/:id", app.deleteCommentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/subscribe/comments", app.handleSubscribeToComments)

	router.HandlerFunc(http.MethodGet, "/signin", app.handleTempAuthTest)
	router.HandlerFunc(http.MethodGet, "/auth", app.handleSignInWithProvider)
	router.HandlerFunc(http.MethodGet, "/auth/callback", app.handleAuthCallback)
	router.HandlerFunc(http.MethodGet, "/signout", app.handleSignOut)

	return app.recoverPanic(app.enableCORS(app.rateLimit(router)))
}
