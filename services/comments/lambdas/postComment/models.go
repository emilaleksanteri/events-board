package main

import (
	"database/sql"
)

type Models struct {
	Comments CommentModel
	Posts    PostModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Comments: CommentModel{DB: db},
		Posts:    PostModel{DB: db},
	}
}
