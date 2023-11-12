package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"errors"

	"math/rand"

	"github.com/emilaleksanteri/pubsub/internal/data"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	CSRF_TOKEN    = "__Secure-events_csrf_token"
	SESSION_TOKEN = "__Secure-events_session_token"
)

// cration fetch user data from provider via goth
// crate user profile
// save provider sent details
// create a pair of session keys

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
	cookies := r.Cookies()
	fmt.Println(cookies)
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

	t, _ := template.New("foo").Parse(userTemplate)
	t.Execute(w, user)
}

func (app *application) handleSignOut(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)
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
	// TODO: replace this
	key    = "randomString"
	MaxAge = 86400 * 30
	IsProd = false
)

func (app *application) initAuth() {
	store := sessions.NewCookieStore([]byte(key))
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
