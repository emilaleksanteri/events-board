package data

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
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
	ParentId         int64      `json:"parent_id"`
	User             *User      `json:"user"`
}

type SqlComment struct {
	Id        sql.NullInt64
	PostId    sql.NullInt64
	Body      sql.NullString
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
}

func parseValidComment(sql *SqlComment, c *Comment) {
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

func (c CommentModel) Insert(comment *Comment, userId int64) error {
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

	comment.User = &user

	return nil
}

func (c CommentModel) Get(id int64, filters *Filters) (*Comment, error) {
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
	args := []any{id, filters.Take, filters.Offset}
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
		var tempUser sqlUser

		err = rows.Scan(
			&tempComment.Id,
			&tempComment.PostId,
			&tempComment.Body,
			&tempComment.CreatedAt,
			&tempComment.UpdatedAt,
			&tempParentId,
			&numSubComments,
			&tempUser.Id,
			&tempUser.Username,
			&tempUser.ProfilePicture,
		)

		if err != nil {
			return nil, err
		}

		tempParentIdInt, err := strconv.ParseInt(tempParentId, 10, 64)
		if err != nil {
			return nil, err
		}

		tempComment.ParentId = tempParentIdInt
		tempComment.NumOfSubComments = numSubComments
		tempComment.User = parseValidUser(&tempUser)

		if tempComment.Id != id {
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

func (c CommentModel) InsertSubComment(comment *Comment, parentId int64, userId int64) error {
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
		&user.Id,
		&user.Username,
		&user.ProfilePicture,
	)

	if err != nil {
		return err
	}

	comment.User = &user

	return nil
}

func (c CommentModel) DeleteComment(id int64) error {
	query := `
	DELETE FROM comments
	WHERE id = $1 OR path <@ $1::text::ltree
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := c.DB.ExecContext(ctx, query, id)
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

func ValidateComment(v *validator.Validator, comment *Comment) {
	v.Check(comment.Body != "", "body", "must be provided")
	v.Check(len(comment.Body) < 20_000, "body", "must cannot be more than 20,000 bytes long")
	v.Check(comment.PostId > 0, "post_id", "must be a positive integer")
}
