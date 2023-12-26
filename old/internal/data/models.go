package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Posts     PostModel
	Comments  CommentModel
	Users     UserModel
	Providers ProviderModel
	Sessions  SessionModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Posts:     PostModel{DB: db},
		Comments:  CommentModel{DB: db},
		Users:     UserModel{DB: db},
		Providers: ProviderModel{DB: db},
		Sessions:  SessionModel{DB: db},
	}
}
