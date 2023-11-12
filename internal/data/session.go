package data

import (
	"time"
)

// fetch session via session token
type Session struct {
	Id        int64     `json:"id"`
	Token     string    `json:"token"`
	UserId    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires"`
}
