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
)

type Post struct {
	Id        int64     `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostMetadata struct {
	LastCommentAt time.Time `json:"last_comment_at"`
	CommentsCount int       `json:"comments_count"`
	LatestComment string    `json:"latest_comment"`
}

type PostData struct {
	Post     Post         `json:"post"`
	Metadata PostMetadata `json:"metadata"`
}

type User struct {
	Id                int64          `json:"id"`
	ProfilePicture    string         `json:"profile_picture"`
	Username          string         `json:"username"`
	sqlId             sql.NullInt64  `json:"-"`
	sqlProfilePicture sql.NullString `json:"-"`
	sqlUsername       sql.NullString `json:"-"`
}

func (u *User) parseSqlNulls() {
	if u.sqlId.Valid {
		u.Id = u.sqlId.Int64
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
	SubComments      []Comment      `json:"sub_comments"`
	Body             string         `json:"body"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	NumOfSubComments int            `json:"num_of_sub_comments"`
	ParentId         int64          `json:"parent_id"`
	User             User           `json:"user"`
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

func (p *PostModel) List(take, skip int) ([]PostData, Metadata, error) {
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

	metadata := Metadata{}
	posts := []PostData{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, query, take, skip)
	if err != nil {
		return nil, metadata, err
	}

	defer rows.Close()

	for rows.Next() {
		postListed := PostData{}
		post := Post{}
		postMetadata := PostMetadata{}
		user := User{}
		commentsCount := 0

		var lastCommentAt sql.NullTime
		var lastCommentBody sql.NullString

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
			return nil, metadata, err
		}

		if lastCommentAt.Valid {
			postMetadata.LastCommentAt = lastCommentAt.Time
		}

		if lastCommentBody.Valid {
			postMetadata.LatestComment = lastCommentBody.String
		}

		postMetadata.CommentsCount = commentsCount
		postListed.Post = post
		postListed.Metadata = postMetadata
		postListed.Post.User = user

		posts = append(posts, postListed)
	}

	if err = rows.Err(); err != nil {
		return nil, metadata, err
	}

	return posts, calculateMetadata(len(posts)), nil
}

func (p *PostModel) Get(id int64, take, offset int) (Post, error) {
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

	post := Post{}
	comments := []Comment{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := p.DB.QueryContext(ctx, query, id, take, offset)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return post, ErrRecordNotFound
		default:
			return post, err
		}
	}

	defer rows.Close()

	for rows.Next() {
		comment := Comment{}
		user := User{}
		commentUser := User{}
		numOfSubComments := 0

		err := rows.Scan(
			&post.Id,
			&post.Body,
			&post.CreatedAt,
			&post.UpdatedAt,
			&comment.sqlId,
			&comment.sqlBody,
			&comment.sqlCreatedAt,
			&comment.sqlUpdatedAt,
			&comment.sqlPostId,
			&numOfSubComments,
			&user.Id,
			&user.Username,
			&user.ProfilePicture,
			&commentUser.sqlId,
			&commentUser.sqlUsername,
			&commentUser.sqlProfilePicture,
		)

		if err != nil {
			return post, err
		}

		post.User = user

		comment.parseSqlNulls()
		commentUser.parseSqlNulls()

		if comment.Id != 0 {
			comment.NumOfSubComments = numOfSubComments
			comment.User = commentUser
			comments = append(comments, comment)
		}
	}

	if err := rows.Err(); err != nil {
		return post, err
	}

	if post.Id == 0 {
		return post, ErrRecordNotFound
	}

	post.Comments = comments
	return post, nil
}
