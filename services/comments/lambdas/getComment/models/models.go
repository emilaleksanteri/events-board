package models

import (
	"context"
	"database/sql"
	"time"
)

type Models struct {
	Comments CommentModel
}

func OpenDB(addr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewModels(db *sql.DB) Models {
	return Models{
		Comments: CommentModel{DB: db},
	}
}
