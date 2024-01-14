package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Like LikeModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Like: LikeModel{DB: db},
	}
}

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
		return err
	}

	return nil
}
