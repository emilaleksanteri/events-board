package main

import (
	"database/sql"
)

type Models struct {
	Posts PostModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Posts: PostModel{DB: db},
	}
}
