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
	Like LikeModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Like: LikeModel{DB: db},
	}
}
