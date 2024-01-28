package main

import (
	"context"
	"database/sql"
	"time"
)

type CommentModel struct {
	DB *sql.DB
}

func (c *CommentModel) updateCommentLikes(commentId int64) (int, int64, error) {
	query := `
		update comments set total_likes = total_likes + 1
		where id = $1
		returning total_likes, user_id
	`

	var likes int
	var userId int64
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, commentId).Scan(&likes, &userId)
	if err != nil {
		return 0, 0, err
	}

	return likes, userId, nil
}
