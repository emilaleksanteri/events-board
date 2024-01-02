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
	Id               int64          `json:"id"`
	PostId           int64          `json:"post_id"`
	SubComments      []*Comment     `json:"sub_comments"`
	Body             string         `json:"body"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	NumOfSubComments int            `json:"num_of_sub_comments"`
	ParentId         int64          `json:"parent_id"`
	User             *User          `json:"user"`
	sqlId            sql.NullInt64  `json:"-"`
	sqlPostId        sql.NullInt64  `json:"-"`
	sqlBody          sql.NullString `json:"-"`
	sqlCreatedAt     sql.NullTime   `json:"-"`
	sqlUpdatedAt     sql.NullTime   `json:"-"`
}

func (c *Comment) parseSqlNulls() {
	if c.sqlId.Valid {
		c.Id = c.sqlId.Int64
	}

	if c.sqlPostId.Valid {
		c.PostId = c.sqlPostId.Int64
	}

	if c.sqlBody.Valid {
		c.Body = c.sqlBody.String
	}

	if c.sqlCreatedAt.Valid {
		c.CreatedAt = c.sqlCreatedAt.Time
	}

	if c.sqlUpdatedAt.Valid {
		c.UpdatedAt = c.sqlUpdatedAt.Time
	}
}

func (c *CommentModel) getComment(commentId int64, take, offset int) (*Comment, error) {
	query := `
	WITH main_comment as (
		SELECT comments.id, comments.post_id, comments.body, comments.created_at, 
		comments.updated_at, comments.path,
		(select count(*) from comments 
		where path = id::text::ltree) as num_of_sub_comments, 
		users.id as comment_user_id, users.username as comment_user_name,
		users.profile_picture as comment_user_profile_picture
		FROM comments
		LEFT JOIN users ON users.id = comments.user_id
		WHERE comments.id = $1
		GROUP BY comments.id, users.id
	),
	sub_comments as (
		SELECT comments.id, comments.post_id, comments.body, comments.created_at, 
		comments.updated_at, comments.path,
		(select count(*) from comments 
		where path = id::text::ltree) as num_of_sub_comments, 
		users.id as sub_user_id, users.username as sub_username,
		users.profile_picture as sub_profile_picture
		FROM comments
		LEFT JOIN users ON users.id = comments.user_id
		WHERE comments.path <@ $1::text::ltree
		GROUP BY comments.id, users.id
		ORDER BY comments.created_at ASC
		LIMIT $2
		OFFSET $3
	)
	SELECT * from main_comment
	UNION ALL
	SELECT * FROM sub_comments
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{commentId, take, offset}
	rows, err := c.DB.QueryContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	defer rows.Close()
	var comment *Comment
	var comments []*Comment

	for rows.Next() {
		tempComment := Comment{}
		tempParentId := ""
		numSubComments := 0
		user := User{}

		err = rows.Scan(
			&tempComment.sqlId,
			&tempComment.sqlPostId,
			&tempComment.sqlBody,
			&tempComment.sqlCreatedAt,
			&tempComment.sqlUpdatedAt,
			&tempParentId,
			&numSubComments,
			&user.sqlId,
			&user.sqlUsername,
			&user.sqlProfilePicture,
		)

		if err != nil {
			return nil, err
		}

		user.parseSqlNulls()
		tempComment.parseSqlNulls()
		tempParentIdInt, err := strconv.ParseInt(tempParentId, 10, 64)
		if err != nil {
			return nil, err
		}

		tempComment.ParentId = tempParentIdInt
		tempComment.NumOfSubComments = numSubComments
		tempComment.User = &user

		if tempComment.Id != commentId {
			comments = append(comments, &tempComment)
		} else {
			comment = &tempComment
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if comment == nil {
		return nil, ErrRecordNotFound
	}

	comment.SubComments = comments
	return comment, nil
}

func (app *app) getCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid comment id parameter"))
		return
	}
	qs := r.URL.Query()
	take, err := app.readInt(qs, "take", 10)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	offset, err := app.readInt(qs, "offset", 0)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	comment, err := app.models.Comments.getComment(commentId, take, offset)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			app.notFoundHandler(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comment": comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
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
	r.Route("/comments", func(r chi.Router) {
		r.Get("/healthcheck", app.healthcheckHandler)
		r.Get("/{id}", app.getCommentHandler)
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
