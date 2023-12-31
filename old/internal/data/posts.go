package data

import (
	"context"
	"database/sql"
	"errors"
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
	User      *User     `json:"user"`
}

type PostMetadata struct {
	LastCommentAt time.Time `json:"last_comment_at"`
	CommentsCount int       `json:"comments_count"`
	LatestComment string    `json:"latest_comment"`
}

type PostData struct {
	Post     *Post         `json:"post"`
	Metadata *PostMetadata `json:"metadata"`
}

func (p PostModel) Insert(post *Post, userId int64) error {
	query := `
	with insert_post as (
		insert into posts (body, user_id)
		values ($1, $2)
		returning id, created_at
	) select insert_post.id, insert_post.created_at, 
	users.id as usr_id, users.username, users.profile_picture from insert_post
	left join users on users.id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var postUser User

	err := p.DB.QueryRowContext(ctx, query, post.Body, userId).Scan(
		&post.Id,
		&post.CreatedAt,
		&postUser.Id,
		&postUser.Username,
		&postUser.ProfilePicture,
	)

	if err != nil {
		return err
	}

	post.User = &postUser
	return nil
}

func (p PostModel) GetAll(filters Filters) ([]*PostData, Metadata, error) {
	query := `
	SELECT post.id, post.body, post.created_at, post.updated_at, 
	COUNT(comment.id) AS comments_count, 
	MAX(comment.created_at) AS last_comment_at, MAX(comment.body) as last_comment_body,
	users.id as user_id, users.username as user_username, users.profile_picture as user_pp
	FROM posts AS post
	LEFT JOIN comments AS comment 
		ON comment.post_id = post.id AND comment.path = '0'
	LEFT JOIN users
		ON users.id = post.user_id	
	GROUP BY post.id, users.id
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
	var posts []*PostData
	numOfPosts := 0

	for rows.Next() {
		var postListed PostData
		var post Post
		var postMetadata PostMetadata
		var lastCommentBody sql.NullString
		var user User

		commentsCount := 0
		var lastCommentAt sql.NullTime

		err := rows.Scan(
			&post.Id,
			&post.Body,
			&post.CreatedAt,
			&post.UpdatedAt,
			&commentsCount,
			&lastCommentAt,
			&lastCommentBody,
			&user.Id,
			&user.Username,
			&user.ProfilePicture,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		if lastCommentAt.Valid {
			timeTemp, err := lastCommentAt.Value()
			if err != nil {
				return nil, Metadata{}, err
			}

			asTime, ok := timeTemp.(time.Time)
			if !ok {
				return nil, Metadata{}, errors.New("could not convert time")
			}

			postMetadata.LastCommentAt = asTime
		}

		if lastCommentBody.Valid {
			postMetadata.LatestComment = lastCommentBody.String
		}

		postMetadata.CommentsCount = commentsCount
		postListed.Post = &post
		postListed.Metadata = &postMetadata
		postListed.Post.User = &user

		posts = append(posts, &postListed)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	numOfPosts = len(posts)
	metadata := calculateMetadata(numOfPosts)

	return posts, metadata, nil
}

func (p PostModel) Get(id int64, filters *Filters) (*Post, error) {
	query := `
	SELECT post.id, post.body, post.created_at, post.updated_at, comment.id, 
	comment.body, comment.created_at, comment.updated_at, comment.post_id,
	(select count(*) 
		from comments 
		where path = comment.id::text::ltree
	) as num_of_sub_comments, users.id as user_id, users.username as user_username, 
	users.profile_picture as user_pp, comment_user.id as comment_user_id,
	comment_user.username as comment_user_username, 
	comment_user.profile_picture as comment_user_pp
	FROM posts as post
	LEFT JOIN comments AS comment 
		ON comment.post_id = post.id AND comment.path = '0'
	LEFT JOIN users 
		ON users.id = post.user_id
	LEFT JOIN users AS comment_user 
		ON comment_user.id = comment.user_id
	WHERE post.id = $1
	GROUP BY post.id, comment.id, users.id, comment_user.id
	ORDER BY post.created_at DESC, comment.created_at ASC
	LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{id, filters.Take, filters.Offset}
	rows, err := p.DB.QueryContext(ctx, query, args...)

	if err != nil {
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
		var comment SqlComment
		var realComment Comment
		numOfSubComments := 0
		var user User
		var commentUser sqlUser

		err := rows.Scan(
			&post.Id,
			&post.Body,
			&post.CreatedAt,
			&post.UpdatedAt,
			&comment.Id,
			&comment.Body,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.PostId,
			&numOfSubComments,
			&user.Id,
			&user.Username,
			&user.ProfilePicture,
			&commentUser.Id,
			&commentUser.Username,
			&commentUser.ProfilePicture,
		)

		if err != nil {
			return nil, err
		}

		parseValidComment(&comment, &realComment)
		if realComment.Id != 0 {
			realComment.NumOfSubComments = numOfSubComments
			realComment.User = parseValidUser(&commentUser)
			comments = append(comments, realComment)
		}
		post.User = &user
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if post.Id == 0 {
		return nil, ErrRecordNotFound
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
