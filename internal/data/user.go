package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/emilaleksanteri/pubsub/internal/validator"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrUserNotFound   = errors.New("user not found")
	AnynomousUser     = &User{}
)

type UserModel struct {
	DB *sql.DB
}

type User struct {
	Id             int64  `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	ProfilePicture string `json:"profile_picture"`
	Username       string `json:"username"`
}

type sqlUser struct {
	Id             sql.NullInt64
	Email          sql.NullString
	Name           sql.NullString
	ProfilePicture sql.NullString
	Username       sql.NullString
}

func (u *User) IsAnynomous() bool {
	return u == AnynomousUser
}

func (um *UserModel) Insert(user *User) error {
	query := `
	INSERT INTO users (email, name, profile_picture, username)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{user.Email, user.Name, user.ProfilePicture, user.Username}
	err := um.DB.QueryRowContext(ctx, query, args...).Scan(&user.Id)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (um *UserModel) GetByEmail(email string) (*User, error) {
	query := `
	SELECT id, email, name, profile_picture, username
	FROM users
	WHERE email = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var tempUser sqlUser

	err := um.DB.QueryRowContext(ctx, query, email).
		Scan(
			&tempUser.Id,
			&tempUser.Email,
			&tempUser.Name,
			&tempUser.ProfilePicture,
			&tempUser.Username,
		)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrUserNotFound
		default:
			return nil, err
		}
	}

	fmt.Println(tempUser)

	return parseValidUser(&tempUser), nil
}

func parseValidUser(user *sqlUser) *User {
	return &User{
		Id:             user.Id.Int64,
		Email:          user.Email.String,
		Name:           user.Name.String,
		ProfilePicture: user.ProfilePicture.String,
		Username:       user.Username.String,
	}
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Email != "", "email", "must be provided")
	v.Check(validator.Matches(user.Email, validator.EmailRX), "email", "must be a valid email address")
	v.Check(user.Username != "", "username", "must be provided")
}
