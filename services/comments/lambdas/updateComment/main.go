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

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	chiLambda         *chiadapter.ChiLambda
)

type CommentModel struct {
	DB *sql.DB
}

type Models struct {
	Comments CommentModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Comments: CommentModel{DB: db},
	}
}

type User struct {
	Id                int64          `json:"id"`
	Email             string         `json:"email"`
	Name              string         `json:"name"`
	ProfilePicture    string         `json:"profile_picture"`
	Username          string         `json:"username"`
	sqlId             sql.NullInt64  `json:"-"`
	sqlEmail          sql.NullString `json:"-"`
	sqlName           sql.NullString `json:"-"`
	sqlProfilePicture sql.NullString `json:"-"`
	sqlUsername       sql.NullString `json:"-"`
}

func (u *User) parseSqlNulls() {
	if u.sqlId.Valid {
		u.Id = u.sqlId.Int64
	}

	if u.sqlEmail.Valid {
		u.Email = u.sqlEmail.String
	}

	if u.sqlName.Valid {
		u.Name = u.sqlName.String
	}

	if u.sqlProfilePicture.Valid {
		u.ProfilePicture = u.sqlProfilePicture.String
	}

	if u.sqlUsername.Valid {
		u.Username = u.sqlUsername.String
	}
}

type Comment struct {
	Id               int64      `json:"id"`
	PostId           int64      `json:"post_id"`
	SubComments      []*Comment `json:"sub_comments"`
	Body             string     `json:"body"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	NumOfSubComments int        `json:"num_of_sub_comments"`
	ParentId         int64      `json:"parent_id"`
	User             *User      `json:"user"`
}

func (c *CommentModel) get(id int64) (*Comment, error) {
	query := `
		select comments.id, comments.body, comments.created_at, comments.updated_at,
		users.id, users.username, users.profile_picture
		from comments
		left join users on users.id = comments.user_id
		where comments.id = $1
	`

	comment := &Comment{}
	user := &User{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&comment.Id,
		&comment.Body,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&user.sqlId,
		&user.sqlUsername,
		&user.sqlProfilePicture,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	user.parseSqlNulls()
	comment.User = user

	return comment, nil
}

func (c *CommentModel) update(comment *Comment) error {
	query := `
		update comments
		set body = $1, updated_at = $2
		where id = $3
		returning updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, comment.Body, time.Now(), comment.Id).Scan(&comment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrRecordNotFound
		}

		return err
	}

	return nil
}

func (app *app) updateCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid comment id parameter"))
		return
	}

	var input struct {
		Body string `json:"body"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	comment, err := app.models.Comments.get(commentId)
	if err != nil {
		switch err {
		case ErrRecordNotFound:
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if input.Body == "" {
		app.badRequestResponse(w, r, errors.New("body can not be empty"))
		return
	}

	comment.Body = input.Body
	err = app.models.Comments.update(comment)
	if err != nil {
		switch err {
		case ErrRecordNotFound:
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
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
		r.Get("/healthcheck", app.healthcheckHandler)
		r.Put("/{id}", app.updateCommentHandler)
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
