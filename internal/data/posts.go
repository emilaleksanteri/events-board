package data

import (
	"context"
	"database/sql"
	"errors"
	"log"
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
	Comments  []Comment `json:"comments"`
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
	LEFT JOIN comments AS comment ON comment.post_id = post.id AND comment.path = '0'
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

func (p PostModel) Get(id int64) (*Post, error) {
	query := `
	SELECT post.id, post.body, post.created_at, post.updated_at, comment.id, comment.body, comment.created_at, comment.updated_at
	FROM posts as post
	LEFT JOIN comments AS comment ON comment.post_id = post.id AND comment.path = '0'
	WHERE post.id = $1
	GROUP BY post.id, comment.id
	ORDER BY post.created_at DESC, comment.created_at ASC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := p.DB.QueryContext(ctx, query, id)

	if err != nil {
		log.Println(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	defer rows.Close()

	var post Post
	var comments []Comment

	for rows.Next() {
		comment := struct {
			Id        sql.NullInt64
			Body      sql.NullString
			CreatedAt sql.NullTime
			UpdatedAt sql.NullTime
		}{}

		var realComment Comment

		err := rows.Scan(
			&post.Id,
			&post.Body,
			&post.CreatedAt,
			&post.UpdatedAt,
			&comment.Id,
			&comment.Body,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if comment.Id.Valid {
			realComment.Id = comment.Id.Int64
		}

		if comment.Body.Valid {
			realComment.Body = comment.Body.String
		}

		if comment.CreatedAt.Valid {
			realComment.CreatedAt = comment.CreatedAt.Time
		}

		if comment.UpdatedAt.Valid {
			realComment.UpdatedAt = comment.UpdatedAt.Time
		}

		if realComment.Id != 0 {
			comments = append(comments, realComment)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	post.Comments = comments
	return &post, nil

}

func (p PostModel) Delete(id int64) error {
	query := `
	DELETE FROM
	posts
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := p.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func ValidPost(v *validator.Validator, post *Post) {
	v.Check(post.Body != "", "body", "must be provided")
	v.Check(len(post.Body) <= 20_000, "body", "must not be more than 20,000 bytes long")
}
