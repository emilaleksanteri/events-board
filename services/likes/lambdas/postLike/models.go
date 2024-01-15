package main

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrAlreadyLiked   = errors.New("user has already liked this")
)

type Models struct {
	Like    LikeModel
	Post    PostModel
	Comment CommentModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Like:    LikeModel{DB: db},
		Post:    PostModel{DB: db},
		Comment: CommentModel{DB: db},
	}
}
