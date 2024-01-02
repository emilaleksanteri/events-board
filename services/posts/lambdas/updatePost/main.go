package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

type PostModel struct {
	DB *sql.DB
}

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	chiLambda         *chiadapter.ChiLambda
)

type Post struct {
	Id        int64     `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      *User     `json:"user"`
}

type User struct {
	Id                int64          `json:"id"`
	Email             string         `json:"email"`
	Name              string         `json:"name"`
	ProfilePicture    string         `json:"profile_picture"`
	Username          string         `json:"username"`
	sqlID             sql.NullInt64  `json:"-"`
	sqlEmail          sql.NullString `json:"-"`
	sqlName           sql.NullString `json:"-"`
	sqlProfilePicture sql.NullString `json:"-"`
	sqlUsername       sql.NullString `json:"-"`
}

func (p *PostModel) Update(post *Post) error {
	query := `
	update posts set
	body = $2
	where id = $1
	returning updated_at
	`

	args := []interface{}{post.Id, post.Body}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, args...).Scan(&post.UpdatedAt)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

func (p *PostModel) Get(id int64) (*Post, error) {
	query := `
	select posts.id, posts.body, posts.created_at, 
	posts.updated_at, users.id, users.profile_picture, users.username 
	from posts
	left join users on users.id = posts.user_id
	where posts.id = $1
	`

	post := Post{}
	postUser := User{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, id).Scan(
		&post.Id,
		&post.Body,
		&post.CreatedAt,
		&post.UpdatedAt,
		&postUser.Id,
		&postUser.ProfilePicture,
		&postUser.Username,
	)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	post.User = &postUser
	return &post, nil
}

func (app *app) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	err := app.writeJSON(w, http.StatusOK, envelope{"status": "available"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *app) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, "resource not found")
}

func (app *app) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		app.notFoundHandler(w, r)
		return
	}

	var input struct {
		Body string `json:"body"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if input.Body == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid body")
		return
	}

	if len(input.Body) > 20_000 {
		app.errorResponse(
			w,
			r,
			http.StatusBadRequest,
			"body too long, max 20_000 characters",
		)
		return
	}

	post, err := app.models.Posts.Get(int64(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	post.Body = input.Body
	err = app.models.Posts.Update(post)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"post": post}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

type Models struct {
	Posts PostModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Posts: PostModel{DB: db},
	}
}

type app struct {
	models Models
}

func openDB() (*sql.DB, error) {
	addr := os.Getenv("DB_ADDRESS")
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func init() {
	db, err := openDB()
	if err != nil {
		panic(err)
	}

	app := app{models: NewModels(db)}
	r := chi.NewRouter()
	r.Route("/update", func(r chi.Router) {
		r.Put("/{id}", app.updatePostHandler)
		r.Get("/healthcheck", app.healthcheckHandler)
	})

	r.NotFound(app.notFoundHandler)
	chiLambda = chiadapter.New(r)
}

func Handler(
	ctx context.Context,
	event events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return chiLambda.ProxyWithContext(ctx, event)
}

func main() {
	lambda.StartWithOptions(Handler, lambda.WithContext(context.Background()))
}
