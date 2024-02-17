package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type User struct {
	Id             int64  `json:"id"`
	Username       string `json:"username"`
	ProfilePicture string `json:"profile_picture"`
}

type UserModel struct {
	DB *sql.DB
}

func (u *UserModel) GetUser(userId int64) (User, error) {
	query := `
		select id, username, profile_picture
		from users
		where id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := u.DB.QueryRowContext(ctx, query, userId)

	var user User
	err := row.Scan(&user.Id, &user.Username, &user.ProfilePicture)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return user, ErrRecordNotFound
		default:
			return user, err
		}
	}

	return user, nil
}
