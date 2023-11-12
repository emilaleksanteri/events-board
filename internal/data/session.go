package data

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"
)

// fetch session via session token, token should be 128 bits long
type Session struct {
	Id        int64     `json:"id"`
	Token     string    `json:"token"`
	UserId    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires"`
}

type SessionModel struct {
	DB *sql.DB
}

var (
	ErrSessionNotFound = errors.New("session not found")
)

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func (sm *SessionModel) Insert(userId int) (string, error) {
	query := `
	INSERT INTO sessions (token, user_id, expires_at)
	VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	token, err := generateRandomString(128)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	args := []any{token, userId, expiresAt}
	err = sm.DB.QueryRowContext(ctx, query, args...).Scan()
	if err != nil {
		return "", err
	}

	return token, nil
}

func (sm *SessionModel) GetByUserId(userId string) (*Session, error) {
	query := `
	SELECT id, token, user_id, expires_at
	FROM sessions
	WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var s Session
	err := sm.DB.QueryRowContext(ctx, query, userId).Scan(&s.Id, &s.Token, &s.UserId, &s.ExpiresAt)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	return &s, nil
}
