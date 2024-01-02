package main

import (
	"context"
	"database/sql"
	"time"
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
