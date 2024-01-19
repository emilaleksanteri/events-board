package main

import (
	"context"
	"database/sql"
	"time"
)

type LikeModel struct {
	DB *sql.DB
}

func (l *LikeModel) removePostLike(postId, userId int64) error {
	query := `
		delete from post_likes
		where post_id = $1 and user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := l.DB.ExecContext(ctx, query, postId, userId)
	if err != nil {
		return err
	}

	affected, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (l *LikeModel) removeCommentLike(commentId, userId int64) error {
	query := `
		delete from comment_likes
		where comment_id = $1 and user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := l.DB.ExecContext(ctx, query, commentId, userId)
	if err != nil {
		return err
	}

	affected, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
