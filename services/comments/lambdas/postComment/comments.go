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

func (c *CommentModel) insertRootComment(comment *Comment, userId int64) error {
	query := `
	with insert_comment as (
		INSERT INTO comments (post_id, body, path, user_id)
		VALUES ($1, $2, '0', $3)
		RETURNING id, created_at, updated_at
	) select insert_comment.id, insert_comment.created_at, insert_comment.updated_at,
	users.id as usr_id, users.username, users.profile_picture from insert_comment
	left join users on users.id = $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := c.DB.QueryRowContext(ctx, query, comment.PostId, comment.Body, userId).Scan(
		&comment.Id,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&user.Id,
		&user.Username,
		&user.ProfilePicture,
	)

	if err != nil {
		return err
	}

	comment.SubComments = []Comment{}
	comment.User = user

	return nil
}

func (c *CommentModel) insertSubComment(comment *Comment, userId, parentId int64) error {
	query := `
	with inseet_comment as (
		INSERT INTO comments (post_id, body, path, user_id)
		VALUES ($1, $2, $3::text::ltree, $4)
		RETURNING id, created_at, updated_at, path::text::bigint
	) select inseet_comment.id, inseet_comment.created_at, inseet_comment.updated_at,
	inseet_comment.path, users.id as usr_id, users.username, users.profile_picture from inseet_comment
	left join users on users.id = $4
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := c.DB.QueryRowContext(
		ctx,
		query,
		comment.PostId,
		comment.Body,
		parentId,
		userId,
	).Scan(
		&comment.Id,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&comment.ParentId,
		&user.Id,
		&user.Username,
		&user.ProfilePicture,
	)

	if err != nil {
		return err
	}

	comment.SubComments = []Comment{}
	comment.User = user

	return nil
}

func (c *CommentModel) getParentCommentUserId(parentId int64) (int64, error) {
	query := `
	select user_id from comments where id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var userId int64

	err := c.DB.QueryRowContext(ctx, query, parentId).Scan(&userId)

	if err != nil {
		return 0, err
	}

	return userId, nil
}
