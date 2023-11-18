package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/emilaleksanteri/pubsub/internal/auth"
)

// fetch session via session token, token should be 128 bits long
type Session struct {
	Id        int64     `json:"id"`
	Token     string    `json:"token"`
	UserId    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires"`
}

type CachedUser struct {
	UserId         int64  `json:"user_id"`
	Username       string `json:"username"`
	ProfilePicture string `json:"profile_picture"`
}

func (cu CachedUser) MarshalBinary() ([]byte, error) {
	return json.Marshal(cu)
}

func (cu *CachedUser) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &cu)
}

type SessionModel struct {
	DB *sql.DB
}

var (
	ErrSessionNotFound = errors.New("session not found")
)

func (sm *SessionModel) Insert(userId int64) (string, error) {
	query := `
	INSERT INTO sessions (token, user_id, expires_at)
	VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	token, err := auth.GenerateToken(128)
	if err != nil {
		return "", err
	}

	stringToken := fmt.Sprintf("%x", token)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	args := []any{stringToken, userId, expiresAt}
	err = sm.DB.QueryRowContext(ctx, query, args...).Scan()
	if err != nil {
		return "", err
	}

	return stringToken, nil
}

func (sm *SessionModel) GetByUserId(userId int64) (string, error) {
	query := `
	SELECT id, token, user_id, expires_at
	FROM sessions
	WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var s Session
	err := sm.DB.QueryRowContext(ctx, query, userId).
		Scan(&s.Id, &s.Token, &s.UserId, &s.ExpiresAt)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return "", ErrSessionNotFound
		}
		return "", err
	}

	return s.Token, nil
}

func (sm *SessionModel) GetByToken(token string) (*Session, error) {
	query := `
	SELECT id, token, user_id, expires_at
	FROM sessions
	WHERE token = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var s Session
	err := sm.DB.QueryRowContext(ctx, query, token).
		Scan(&s.Id, &s.Token, &s.UserId, &s.ExpiresAt)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	return &s, nil
}
