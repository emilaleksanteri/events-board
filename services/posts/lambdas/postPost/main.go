package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	chiLambda         *chiadapter.ChiLambda
)

type PostModel struct {
	DB *sql.DB
}

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

func (p PostModel) Insert(post *Post, userId int64) error {
	query := `
	with insert_post as (
		insert into posts (body, user_id)
		values ($1, $2)
		returning id, created_at
	) select insert_post.id, insert_post.created_at, 
	users.id as usr_id, users.username, users.profile_picture from insert_post
	left join users on users.id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var postUser User

	err := p.DB.QueryRowContext(ctx, query, post.Body, userId).Scan(
		&post.Id,
		&post.CreatedAt,
		&postUser.Id,
		&postUser.Username,
		&postUser.ProfilePicture,
	)

	if err != nil {
		return err
	}

	post.User = &postUser
	return nil
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

func (app *app) createHandler(w http.ResponseWriter, r *http.Request) {
	tempUsrId := int64(2)
	var input struct {
		Body string `json:"body"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if input.Body == "" {
		app.errorResponse(
			w,
			r,
			http.StatusBadRequest,
			"missing body, min length is 1 character",
		)
		return
	}

	if len(input.Body) > 20_000 {
		app.errorResponse(
			w,
			r,
			http.StatusBadRequest,
			"body too long, max is 20_000 characters",
		)
		return
	}

	post := &Post{
		Body: input.Body,
	}

	err = app.models.Posts.Insert(post, tempUsrId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/posts/%d", post.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"post": post}, headers)
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
	r.Get("/create/healthcheck", app.healthcheckHandler)
	r.Post("/create", app.createHandler)
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
