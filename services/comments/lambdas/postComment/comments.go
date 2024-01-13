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
		&user.sqlId,
		&user.sqlUsername,
		&user.sqlProfilePicture,
	)

	if err != nil {
		return err
	}

	user.parseSqlNulls()
	comment.User = &user

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

	args := []any{comment.PostId, comment.Body, parentId, userId}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := c.DB.QueryRowContext(ctx, query, args...).Scan(
		&comment.Id,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&comment.ParentId,
		&user.sqlId,
		&user.sqlUsername,
		&user.sqlProfilePicture,
	)

	if err != nil {
		return err
	}

	user.parseSqlNulls()
	comment.User = &user

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
