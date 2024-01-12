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
