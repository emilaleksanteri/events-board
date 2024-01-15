package main

import (
	"context"
	"database/sql"
	"time"
)

type PostModel struct {
	DB *sql.DB
}

func (p *PostModel) updatePostLikes(postId int64) error {
	query := `
		update posts set total_likes = total_likes + 1
		where id = $1
		returning total_likes
	`

	var likes int64
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, postId).Scan(&likes)
	if err != nil {
		return err
	}

	return nil
}
