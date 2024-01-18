package main

import (
	"database/sql"
	"errors"
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
