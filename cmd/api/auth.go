package main

import (
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/emilaleksanteri/pubsub/internal/auth"
	"github.com/emilaleksanteri/pubsub/internal/data"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	CSRF_COOKIE    = "__Secure-events_csrf_token"
	SESSION_COOKIE = "__Secure-events_session_token"
)

func randomUserAdjectiveThing() string {
	possibleOnes := []string{
		"beloved",
		"adored",
		"cherished",
		"treasured",
		"prized",
		"favorite",
		"precious",
		"favorite",
		"coolest",
		"best",
	}
	return possibleOnes[rand.Intn(len(possibleOnes))]
}

func (app *application) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	userInDb, err := app.models.Users.GetByEmail(user.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUserNotFound):
			userInDb = &data.User{
				Email:          user.Email,
				Name:           user.Name,
				ProfilePicture: user.AvatarURL,
				Username:       fmt.Sprintf("%s-user", randomUserAdjectiveThing()),
			}

			err = app.models.Users.Insert(userInDb)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	_, err = app.models.Providers.GetByUser(userInDb.Id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrProviderNotFound):
			providerUser := &data.Provider{UserId: userInDb.Id}
			err = app.models.Providers.Insert(providerUser)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	sessionToken, err := app.models.Sessions.GetByUserId(userInDb.Id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrSessionNotFound):
			sessionToken, err = app.models.Sessions.Insert(userInDb.Id)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	csrf := auth.MakeToken(
		fmt.Sprintf("%s%s",
			sessionToken,
			os.Getenv("SESSION_SECRET"),
		),
	)

	expiry := time.Now().AddDate(0, 1, 0)
	maxAge := 86400 * 30
	app.SetSecureCookie(w, SESSION_COOKIE, sessionToken, expiry, maxAge)
	app.SetSecureCookie(w, CSRF_COOKIE, csrf, expiry, maxAge)

	t, _ := template.New("foo").Parse(userTemplate)
	t.Execute(w, user)
}

func (app *application) handleSignOut(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)
	app.DeleteSecureCookie(w, SESSION_COOKIE)
	app.DeleteSecureCookie(w, CSRF_COOKIE)

	w.Header().Set("Location", "/signin")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (app *application) handleSignInWithProvider(w http.ResponseWriter, r *http.Request) {
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		t, _ := template.New("foo").Parse(userTemplate)
		t.Execute(w, gothUser)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func (app *application) handleTempAuthTest(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("foo").Parse(indexTemplate)
	t.Execute(w, nil)
}

const (
	key    = "randomString"
	MaxAge = 86400 * 30
	IsProd = false
)

var Key = os.Getenv("SESSION_SECRET")

func (app *application) initAuth() {
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.MaxAge(MaxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = IsProd

	gothic.Store = store

	goth.UseProviders(
		google.New(os.Getenv("PUBSUB_GOOGLE_CLIENT_ID"), os.Getenv("PUBSUB_GOOGLE_CLIENT_SECRET"), "http://localhost:4000/auth/callback?provider=google"),
	)
}

var indexTemplate = `
    <p><a href="/auth?provider=google">Log in with Google</a></p>`

var userTemplate = `
<p><a href="/signout?provider={{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`
