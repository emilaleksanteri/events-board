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

func (c *CommentModel) getComment(commentId int64, take, offset int) (Comment, error) {
	query := `
	WITH main_comment as (
		SELECT comments.id, comments.post_id, comments.body, comments.created_at, 
		comments.updated_at, comments.path,
		(select count(*) from comments as c
		where path = comments.id::text::ltree) as num_of_sub_comments, 
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
		(select count(*) from comments as c
		where path = comments.id::text::ltree) as num_of_sub_comments, 
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
	comment := Comment{}
	comments := []Comment{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, commentId, take, offset)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return comment, ErrRecordNotFound
		default:
			return comment, err
		}
	}

	defer rows.Close()

	for rows.Next() {
		tempComment := Comment{}
		user := User{}

		tempParentId := ""
		numSubComments := 0

		err = rows.Scan(
			&tempComment.Id,
			&tempComment.PostId,
			&tempComment.Body,
			&tempComment.CreatedAt,
			&tempComment.UpdatedAt,
			&tempParentId,
			&numSubComments,
			&user.Id,
			&user.Username,
			&user.ProfilePicture,
		)

		if err != nil {
			return comment, err
		}

		tempParentIdInt, err := strconv.ParseInt(tempParentId, 10, 64)
		if err != nil {
			return comment, err
		}

		tempComment.ParentId = tempParentIdInt
		tempComment.NumOfSubComments = numSubComments
		tempComment.User = user

		if tempComment.Id != commentId {
			tempComment.SubComments = []Comment{}
			comments = append(comments, tempComment)
		} else {
			comment = tempComment
		}
	}

	if err = rows.Err(); err != nil {
		return comment, err
	}

	if comment.Id == 0 {
		return comment, ErrRecordNotFound
	}

	comment.SubComments = comments
	return comment, nil
}
