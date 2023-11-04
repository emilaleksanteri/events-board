package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/emilaleksanteri/pubsub/internal/validator"
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

func (p PostModel) GetAll(filters Filters) ([]*Post, Metadata, error) {
	// query get num of comments and most recent comment
	query := `
	SELECT post.id, post.body, post.created_at, post.updated_at, COUNT(comment.id) AS comments_count, MAX(comment.created_at) AS last_comment_at
	FROM posts AS post
	LEFT JOIN comments AS comment ON comment.post_id = post.id
	GROUP BY post.id
	ORDER BY post.created_at DESC
	LIMIT $1 OFFSET $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{filters.Take, filters.Offset}
	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	posts := []*Post{}
	numOfPosts := 0

	for rows.Next() {
		var post Post
		commentsCount := 0
		var lastCommentAt any

		err := rows.Scan(
			&post.Id,
			&post.Body,
			&post.CreatedAt,
			&post.UpdatedAt,
			&commentsCount,
			&lastCommentAt,
		)

		if lastCommentAt != nil {
			lastCommentAt = lastCommentAt.(time.Time)
		}

		if err != nil {
			return nil, Metadata{}, err
		}

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	numOfPosts = len(posts)
	metadata := calculateMetadata(numOfPosts)

	return posts, metadata, nil

}

func ValidPost(v *validator.Validator, post *Post) {
	v.Check(post.Body != "", "body", "must be provided")
	v.Check(len(post.Body) <= 20_000, "body", "must not be more than 20,000 bytes long")
}
