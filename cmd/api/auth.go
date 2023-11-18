package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/emilaleksanteri/pubsub/internal/auth"
	"github.com/emilaleksanteri/pubsub/internal/data"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/redis/go-redis/v9"
)

const (
	CSRF_COOKIE           = "__Secure-events_csrf_token"
	SESSION_COOKIE        = "__Secure-events_session_token"
	CLIENT_REDIRECT_COKIE = "__events_client_redirect"
	MaxAge                = 60 * 60 * 24 * 30
	IsProd                = false
)

var (
	AuthKey = os.Getenv("SESSION_SECRET")
	Enckey  string
)

func (app *application) initAuth() sessions.Store {
	Enckey, err := auth.GenerateToken(16)
	if err != nil {
		app.logger.Error(err.Error())
		os.Exit(1)
	}

	store := sessions.NewCookieStore([]byte(AuthKey), []byte(Enckey))
	store.MaxAge(MaxAge)
	store.Options.Path = "/"
	store.Options.MaxAge = MaxAge
	store.Options.HttpOnly = true
	store.Options.Secure = IsProd

	gothic.Store = store
	sessionStore := gothic.Store

	goth.UseProviders(
		google.New(os.Getenv("PUBSUB_GOOGLE_CLIENT_ID"), os.Getenv("PUBSUB_GOOGLE_CLIENT_SECRET"), "http://localhost:4000/auth/callback?provider=google"),
	)

	return sessionStore
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
				Username: fmt.Sprintf("%s-user",
					data.RandomUserAdjectiveThing(),
				),
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
			providerUser := data.Provider{
				UserId:            userInDb.Id,
				Provider:          user.Provider,
				AccessToken:       user.AccessToken,
				AccessTokenSecret: user.AccessTokenSecret,
				ExpiresAt:         user.ExpiresAt,
				IdToken:           user.IDToken,
				RefreshToken:      user.RefreshToken,
			}

			err = app.models.Providers.Insert(&providerUser)
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

	userRedis := data.CachedUser{
		UserId:         userInDb.Id,
		Username:       userInDb.Username,
		ProfilePicture: userInDb.ProfilePicture,
	}

	err = app.redis.Set(r.Context(), sessionToken, userRedis, 30*24*time.Hour).Err()

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	csrf := auth.MakeToken(fmt.Sprintf("%s%s", sessionToken, AuthKey))
	expiry := time.Now().AddDate(0, 1, 0)

	redirectCookie := FindCookie(r, CLIENT_REDIRECT_COKIE)
	if redirectCookie == nil {
		app.noProvidedAuthRedirectUrl(w, r)
		return
	}

	app.SetSecureCookie(w, SESSION_COOKIE, sessionToken, expiry, MaxAge)
	app.SetSecureCookie(w, CSRF_COOKIE, csrf, expiry, MaxAge)

	w.Header().Set("Location", redirectCookie.Value)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (app *application) handleSignOut(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)
	sessionCookie := FindCookie(r, SESSION_COOKIE)
	if sessionCookie != nil {
		err := app.redis.Del(r.Context(), sessionCookie.Value).Err()
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		app.DeleteSecureCookie(w, SESSION_COOKIE)
		app.DeleteSecureCookie(w, CSRF_COOKIE)
	}
	w.Header().Set("Location", "/signin")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (app *application) handleSignInWithProvider(w http.ResponseWriter, r *http.Request) {
	sessionCookie := FindCookie(r, SESSION_COOKIE)

	redirectCookie := FindCookie(r, CLIENT_REDIRECT_COKIE)
	if redirectCookie == nil {
		app.noProvidedAuthRedirectUrl(w, r)
		return
	}

	if sessionCookie != nil {
		var userRedis data.CachedUser
		err := app.redis.Get(r.Context(), sessionCookie.Value).Scan(&userRedis)
		if err != nil {
			switch {
			case errors.Is(err, redis.Nil):
			default:
				app.serverErrorResponse(w, r, err)
				return
			}
		}

		if userRedis.UserId != 0 {
			w.Header().Set("Location", redirectCookie.Value)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}

	gothic.BeginAuthHandler(w, r)
}

func (app *application) getUserSession(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnynomous() {
		w.Header().Set("Location", "/signin")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	t, _ := template.New("foo").Parse(authTemplate)
	t.Execute(w, user)

}

func (app *application) handleTempAuthTest(w http.ResponseWriter, r *http.Request) {
	redirectUrl := r.URL.Query().Get("redirect")
	if redirectUrl == "" {
		app.noProvidedAuthRedirectUrl(w, r)
		return
	}

	app.SetSecureCookie(
		w,
		CLIENT_REDIRECT_COKIE,
		redirectUrl,
		time.Now().Add(10*time.Minute),
		MaxAge,
	)

	t, _ := template.New("foo").Parse(indexTemplate)
	t.Execute(w, nil)
}

var indexTemplate = `
    <p><a href="/auth?provider=google">Log in with Google</a></p>`

var authTemplate = `
<p><a href="/signout?provider=google">logout</a></p>
<p>username: {{.Username}} is logged in</p>
`

var userTemplate = `
<p><a href="/signout?provider={{.Provider}}">logout</a></p>
<p><a href="/profile">profile</a></p>
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
