package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type PostModel struct {
	DB *sql.DB
}

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

type Models struct {
	Posts PostModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Posts: PostModel{DB: db},
	}
}
