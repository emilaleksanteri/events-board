package main

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type LikeModel struct {
	DB *sql.DB
}

type PostLike struct {
	Id         int64     `json:"id"`
	PostId     int64     `json:"post_id"`
	UserId     int64     `json:"user_id"`
	Created_at time.Time `json:"created_at"`
}

type CommentLike struct {
	Id         int64     `json:"id"`
	CommentId  int64     `json:"comment_id"`
	UserId     int64     `json:"user_id"`
	Created_at time.Time `json:"created_at"`
}

func (l *LikeModel) likePost(postLike *PostLike) error {
	query := `
		insert into post_likes (post_id, user_id)
		values ($1, $2)
		returning id, created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{postLike.PostId, postLike.UserId}
	err := l.DB.QueryRowContext(ctx, query, args...).Scan(&postLike.Id, &postLike.Created_at)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "duplicate key value violates unique constraint"):
			return ErrAlreadyLiked
		default:
			return err
		}
	}

	return nil
}

func (l *LikeModel) undoPostLike(postLike *PostLike) error {
	query := `
		delete from post_likes
		where id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := l.DB.ExecContext(ctx, query, postLike.Id)
	if err != nil {
		return err
	}

	return nil
}

func (l *LikeModel) likeComment(commentLike *CommentLike) error {
	query := `
		insert into comment_likes (comment_id, user_id)
		values ($1, $2)
		returning id, created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{commentLike.CommentId, commentLike.UserId}
	err := l.DB.QueryRowContext(ctx, query, args...).Scan(&commentLike.Id, &commentLike.Created_at)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "duplicate key value violates unique constraint"):
			return ErrAlreadyLiked
		default:
			return err
		}
	}

	return nil
}

func (l *LikeModel) undoCommentLike(commentLike *CommentLike) error {
	query := `
		delete from comment_likes
		where id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := l.DB.ExecContext(ctx, query, commentLike.Id)
	if err != nil {
		return err
	}

	return nil
}
