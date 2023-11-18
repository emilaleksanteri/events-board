package main

import (
	"context"
	"net/http"

	"github.com/emilaleksanteri/pubsub/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.CachedUser) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.CachedUser {
	user, ok := r.Context().Value(userContextKey).(*data.CachedUser)
	if !ok {
		panic("missing user value in req context!!")
	}

	return user
}
