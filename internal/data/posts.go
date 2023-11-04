package data

import (
	"context"
	"database/sql"
	"github.com/emilaleksanteri/pubsub/internal/validator"
	"time"
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

func (p PostModel) Insert(post *Post) error {
	query := `
	INSERT INTO posts (body)
	VALUES ($1)
	RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.DB.QueryRowContext(ctx, query, post.Body).
		Scan(&post.Id, &post.CreatedAt, &post.UpdatedAt)
}

func ValidPost(v *validator.Validator, post *Post) {
	v.Check(post.Body != "", "body", "must be provided")
	v.Check(len(post.Body) <= 20_000, "body", "must not be more than 20,000 bytes long")
}
