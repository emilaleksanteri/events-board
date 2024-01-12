package main

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
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

func (c *CommentModel) getComment(commentId int64, take, offset int) (*Comment, error) {
	query := `
	WITH main_comment as (
		SELECT comments.id, comments.post_id, comments.body, comments.created_at, 
		comments.updated_at, comments.path,
		(select count(*) from comments 
		where path = id::text::ltree) as num_of_sub_comments, 
		users.id as comment_user_id, users.username as comment_user_name,
		users.profile_picture as comment_user_profile_picture
		FROM comments
		LEFT JOIN users ON users.id = comments.user_id
		WHERE comments.id = $1
		GROUP BY comments.id, users.id
	),
	sub_comments as (
		SELECT comments.id, comments.post_id, comments.body, comments.created_at, 
		comments.updated_at, comments.path,
		(select count(*) from comments 
		where path = id::text::ltree) as num_of_sub_comments, 
		users.id as sub_user_id, users.username as sub_username,
		users.profile_picture as sub_profile_picture
		FROM comments
		LEFT JOIN users ON users.id = comments.user_id
		WHERE comments.path <@ $1::text::ltree
		GROUP BY comments.id, users.id
		ORDER BY comments.created_at ASC
		LIMIT $2
		OFFSET $3
	)
	SELECT * from main_comment
	UNION ALL
	SELECT * FROM sub_comments
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{commentId, take, offset}
	rows, err := c.DB.QueryContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	defer rows.Close()
	var comment *Comment
	var comments []*Comment

	for rows.Next() {
		tempComment := Comment{}
		tempParentId := ""
		numSubComments := 0
		user := User{}

		err = rows.Scan(
			&tempComment.sqlId,
			&tempComment.sqlPostId,
			&tempComment.sqlBody,
			&tempComment.sqlCreatedAt,
			&tempComment.sqlUpdatedAt,
			&tempParentId,
			&numSubComments,
			&user.sqlId,
			&user.sqlUsername,
			&user.sqlProfilePicture,
		)

		if err != nil {
			return nil, err
		}

		user.parseSqlNulls()
		tempComment.parseSqlNulls()
		tempParentIdInt, err := strconv.ParseInt(tempParentId, 10, 64)
		if err != nil {
			return nil, err
		}

		tempComment.ParentId = tempParentIdInt
		tempComment.NumOfSubComments = numSubComments
		tempComment.User = &user

		if tempComment.Id != commentId {
			comments = append(comments, &tempComment)
		} else {
			comment = &tempComment
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if comment == nil {
		return nil, ErrRecordNotFound
	}

	comment.SubComments = comments
	return comment, nil
}
