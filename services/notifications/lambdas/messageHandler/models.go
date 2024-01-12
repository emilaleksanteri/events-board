package main

import (
	"database/sql"
)

type Models struct {
	SocialConns SocialConnsModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		SocialConns: SocialConnsModel{DB: db},
	}
}
