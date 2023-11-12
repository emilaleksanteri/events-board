package data

import (
	"time"
)

// check user actually has an account with us, also in case we need id token data
// id token is rs256 jwt token
type Provider struct {
	Id                int64     `json:"id"`
	Provider          string    `json:"provider"`
	AccessToken       string    `json:"access_token"`
	RefreshToken      string    `json:"refresh_token"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	UserId            int64     `json:"user_id"`
	IdToken           string    `json:"id_token"`
	AccessTokenSecret string    `json:"access_token_secret"`
}
