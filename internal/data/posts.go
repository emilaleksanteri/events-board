package data

import (
	"context"
	"database/sql"
	"errors"
	"log"
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

func (p PostModel) Insert(post *Post) error {
	query := `
	INSERT INTO posts (body)
	VALUES ($1)
	RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.DB.QueryRowContext(ctx, query, post.Body).
		Scan(&post.Id, &post.CreatedAt, &post.UpdatedAt)
}

func (p PostModel) GetAll(filters Filters) ([]*PostData, Metadata, error) {
	query := `
	SELECT post.id, post.body, post.created_at, post.updated_at, 
	COUNT(comment.id) AS comments_count, 
	MAX(comment.created_at) AS last_comment_at, MAX(comment.body) as last_comment_body
	FROM posts AS post
	LEFT JOIN comments AS comment ON comment.post_id = post.id AND comment.path = '0'
	GROUP BY post.id
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
	) as num_of_sub_comments
	FROM posts as post
	LEFT JOIN comments AS comment ON comment.post_id = post.id AND comment.path = '0'
	WHERE post.id = $1
	GROUP BY post.id, comment.id
	ORDER BY post.created_at DESC, comment.created_at ASC
	LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{id, filters.Take, filters.Offset}
	rows, err := p.DB.QueryContext(ctx, query, args...)

	if err != nil {
		log.Println(err)
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
		)

		if err != nil {
			return nil, err
		}

		getValidComment(&comment, &realComment)
		if realComment.Id != 0 {
			realComment.NumOfSubComments = numOfSubComments
			comments = append(comments, realComment)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
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
