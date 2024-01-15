package main

import (
	"context"
	"database/sql"
	"time"
)

type CommentModel struct {
	DB *sql.DB
}

func (c *CommentModel) updateCommentLikes(commentId int64) (int, error) {
	query := `
		update comments set total_likes = total_likes + 1
		where id = $1
		returning total_likes
	`

	var likes int
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, commentId).Scan(&likes)
	if err != nil {
		return 0, err
	}

	return likes, nil
}
