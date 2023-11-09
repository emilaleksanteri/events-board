package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/emilaleksanteri/pubsub/internal/validator"
)

type CommentModel struct {
	DB *sql.DB
}

type Comment struct {
	Id               int64      `json:"id"`
	PostId           int64      `json:"post_id"`
	SubComments      []*Comment `json:"sub_comments"`
	Body             string     `json:"body"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	NumOfSubComments int        `json:"num_of_sub_comments"`
}

type SqlComment struct {
	Id        sql.NullInt64
	PostId    sql.NullInt64
	Body      sql.NullString
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
}

func getValidComment(sql *SqlComment, c *Comment) {
	if sql.Id.Valid {
		c.Id = sql.Id.Int64
	}

	if sql.Body.Valid {
		c.Body = sql.Body.String
	}

	if sql.CreatedAt.Valid {
		c.CreatedAt = sql.CreatedAt.Time
	}

	if sql.UpdatedAt.Valid {
		c.UpdatedAt = sql.UpdatedAt.Time
	}

	if sql.PostId.Valid {
		c.PostId = sql.PostId.Int64
	}
}

func (c CommentModel) Insert(comment *Comment) error {
	query := `
	INSERT INTO comments (post_id, body, path)
	VALUES ($1, $2, '0')
	RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, comment.PostId, comment.Body).
		Scan(&comment.Id, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (c CommentModel) Get(id int64) (*Comment, error) {
	query := `
	SELECT id, post_id, body, created_at, updated_at
	FROM comments
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var comment Comment
	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&comment.Id,
		&comment.PostId,
		&comment.Body,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	query = `
	SELECT id, post_id, body, created_at, updated_at
	FROM comments
	WHERE path <@ '$1'
	ORDER BY created_at ASC
	`

	rows, err := c.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	comments := []*Comment{}

	for rows.Next() {
		var comment Comment

		err := rows.Scan(
			&comment.Id,
			&comment.PostId,
			&comment.Body,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	comment.SubComments = comments
	return &comment, nil
}

func (c CommentModel) InsertSubComment(comment *Comment, parentId int64) error {
	query := `
	INSERT INTO comments (post_id, body, path)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at
	`

	args := []any{comment.PostId, comment.Body, fmt.Sprintf("%v", parentId)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, args...).Scan(
		&comment.Id,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func ValidateComment(v *validator.Validator, comment *Comment) {
	v.Check(comment.Body != "", "body", "must be provided")
	v.Check(len(comment.Body) < 20_000, "body", "must cannot be more than 20,000 bytes long")
	v.Check(comment.PostId > 0, "post_id", "must be a positive integer")
}
