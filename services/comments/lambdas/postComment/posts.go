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

func (p *PostModel) GetPostUserId(postId int64) (int64, error) {
	query := `
		select user_id from posts where id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var userId int64
	err := p.DB.QueryRowContext(ctx, query, postId).Scan(&userId)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return 0, ErrRecordNotFound
		default:
			return 0, err
		}
	}

	return userId, nil
}
