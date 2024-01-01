package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type PostModel struct {
	DB *sql.DB
}

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

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

type User struct {
	Id                int64          `json:"id"`
	Email             string         `json:"email"`
	Name              string         `json:"name"`
	ProfilePicture    string         `json:"profile_picture"`
	Username          string         `json:"username"`
	sqlID             sql.NullInt64  `json:"-"`
	sqlEmail          sql.NullString `json:"-"`
	sqlName           sql.NullString `json:"-"`
	sqlProfilePicture sql.NullString `json:"-"`
	sqlUsername       sql.NullString `json:"-"`
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

type Metadata struct {
	PageSize int `json:"page_size"`
}

func calculateMetadata(pageSize int) Metadata {
	if pageSize == 0 {
		return Metadata{}
	}

	return Metadata{
		PageSize: pageSize,
	}
}

func (p *PostModel) List(take, skip int) ([]*PostData, Metadata, error) {
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
	args := []any{take, skip}
	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	var posts []*PostData
	numOfPosts := 0

	for rows.Next() {
		postListed := PostData{}
		post := Post{}
		postMetadata := PostMetadata{}
		var lastCommentBody sql.NullString
		user := User{}

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
			&user.sqlID,
			&user.sqlUsername,
			&user.sqlProfilePicture,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		if lastCommentAt.Valid {
			postMetadata.LastCommentAt = lastCommentAt.Time
		}

		if lastCommentBody.Valid {
			postMetadata.LatestComment = lastCommentBody.String
		}

		if user.sqlID.Valid {
			user.Id = user.sqlID.Int64
		}

		if user.sqlUsername.Valid {
			user.Username = user.sqlUsername.String
		}

		if user.sqlProfilePicture.Valid {
			user.ProfilePicture = user.sqlProfilePicture.String
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

func (p *PostModel) Get(id int64) (*Post, error) {
	query := `
		SELECT id, body, created_at, updated_at
		FROM posts
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []interface{}{id}
	row := p.DB.QueryRowContext(ctx, query, args...)

	var post Post
	err := row.Scan(&post.Id, &post.Body, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		} else {
			return nil, err
		}
	}

	return &post, nil
}

type Models struct {
	Posts PostModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Posts: PostModel{DB: db},
	}
}
