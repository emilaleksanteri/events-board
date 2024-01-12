package main

import (
	"context"
	"database/sql"
	"time"
)

type PostModel struct {
	DB *sql.DB
}

type Post struct {
	Id        int64     `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserId    int64     `json:"user_id"`
}

// returns Post struct with just post id and user_id fields
func (p *PostModel) GetPostWithUser(postId int64) (*Post, error) {
	query := `
		select id, user_id from posts where id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var post Post
	err := p.DB.QueryRowContext(ctx, query, postId).Scan(&post.Id, &post.UserId)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}
