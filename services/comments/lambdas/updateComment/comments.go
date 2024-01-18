package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type CommentModel struct {
	DB *sql.DB
}

type User struct {
	Id             int64  `json:"id"`
	ProfilePicture string `json:"profile_picture"`
	Username       string `json:"username"`
}

type Comment struct {
	Id               int64     `json:"id"`
	PostId           int64     `json:"post_id"`
	SubComments      []Comment `json:"sub_comments"`
	Body             string    `json:"body"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	NumOfSubComments int       `json:"num_of_sub_comments"`
	ParentId         int64     `json:"parent_id"`
	User             User      `json:"user"`
}

func (c *CommentModel) get(id int64) (Comment, error) {
	query := `
		select comments.id, comments.body, comments.created_at, comments.updated_at,
		comments.post_id, comments.path, users.id, 
		users.username, users.profile_picture
		from comments
		left join users on users.id = comments.user_id
		where comments.id = $1
	`

	comment := Comment{}
	user := User{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&comment.Id,
		&comment.Body,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&comment.PostId,
		&comment.ParentId,
		&user.Id,
		&user.Username,
		&user.ProfilePicture,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return comment, ErrRecordNotFound
		}

		return comment, err
	}

	comment.SubComments = []Comment{}
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
