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
	User      *User     `json:"user"`
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

func (p *PostModel) Update(post *Post) error {
	query := `
	update posts set
	body = $2
	where id = $1
	returning updated_at
	`

	args := []interface{}{post.Id, post.Body}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, args...).Scan(&post.UpdatedAt)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

func (p *PostModel) Get(id int64) (*Post, error) {
	query := `
	select posts.id, posts.body, posts.created_at, 
	posts.updated_at, users.id, users.profile_picture, users.username 
	from posts
	left join users on users.id = posts.user_id
	where posts.id = $1
	`

	post := Post{}
	postUser := User{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, id).Scan(
		&post.Id,
		&post.Body,
		&post.CreatedAt,
		&post.UpdatedAt,
		&postUser.Id,
		&postUser.ProfilePicture,
		&postUser.Username,
	)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	post.User = &postUser
	return &post, nil
}
