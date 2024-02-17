package main

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound   = errors.New("record not found")
	ErrAlreadyLiked     = errors.New("user has already liked this")
	ErrAlreadyFollowing = errors.New("user has already followed this")
)

type Models struct {
	Social SocialModel
	User   UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Social: SocialModel{DB: db},
		User:   UserModel{DB: db},
	}
}
