package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type PostModel struct {
	DB *sql.DB
}

type Post struct {
	Id        int64     `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *PostModel) list(take, skip int) (*[]Post, error) {
	query := `
		SELECT id, body, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		OFFSET $1
		LIMIT $2
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{skip, take}
	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := []Post{}
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.Id, &post.Body, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &posts, nil
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

func (app *app) readInt(res events.APIGatewayProxyRequest, key string, defaultValue int) (int, error) {
	s := res.QueryStringParameters[key]
	if s == "" {
		return defaultValue, nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue, errors.New("key must be a valid int")
	}
	return i, nil
}

func (app *app) listPostsHandler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	db, err := openDB()
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("\"%s\"", err.Error()),
		}, nil
	}
	defer db.Close()

	app.models = NewModels(db)
	take, err := app.readInt(event, "take", 10)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf("\"%s\"", err.Error()),
		}, nil
	}

	skip, err := app.readInt(event, "skip", 0)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf("\"%s\"", err.Error()),
		}, nil
	}

	posts, err := app.models.Posts.list(take, skip)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("\"%s\"", err.Error()),
		}, nil
	}

	res, err := json.Marshal(posts)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("\"%s\"", err.Error()),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(res),
	}, nil

}

func (app *app) handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch event.Path {
	case "/":
		response := events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "\"hello from lambda!\"",
		}
		return response, nil
	case "/health":
		response := events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "\"running healthy :)\"",
		}

		return response, nil

	case "/posts":
		return app.listPostsHandler(ctx, event)

	case "/test":
		variable := os.Getenv("DB_ADDRESS")
		response := events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       fmt.Sprintf("\"my db addr is: %s\"", variable),
		}
		return response, nil

	default:
		response := events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "\"not found\"",
		}
		return response, nil
	}
}

func main() {
	app := app{}

	lambda.Start(app.handler)
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
