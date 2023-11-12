package data

import (
	"context"
	"database/sql"
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

type sqlProvider struct {
	Id                sql.NullInt64
	Provider          sql.NullString
	AccessToken       sql.NullString
	RefreshToken      sql.NullString
	CreatedAt         sql.NullTime
	ExpiresAt         sql.NullTime
	UserId            sql.NullInt64
	IdToken           sql.NullString
	AccessTokenSecret sql.NullString
}

type ProviderModel struct {
	DB *sql.DB
}

func validateSqlProvider(p *sqlProvider) *Provider {
	return &Provider{
		Id:                p.Id.Int64,
		Provider:          p.Provider.String,
		AccessToken:       p.AccessToken.String,
		RefreshToken:      p.RefreshToken.String,
		ExpiresAt:         p.ExpiresAt.Time,
		UserId:            p.UserId.Int64,
		IdToken:           p.IdToken.String,
		AccessTokenSecret: p.AccessTokenSecret.String,
	}
}

func (pm *ProviderModel) Insert(p *Provider) error {
	query := `
	INSERT INTO providers (provider, access_token, refresh_token, expires_at, user_id, id_token, access_token_secret)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		p.Provider,
		p.AccessToken,
		p.RefreshToken,
		p.ExpiresAt,
		p.UserId,
		p.IdToken,
		p.AccessTokenSecret,
	}

	return pm.DB.QueryRowContext(ctx, query, args...).Scan(&p.Id)
}

func (pm *ProviderModel) GetByUser(userId string) (*Provider, error) {
	query := `
	SELECT id, provider, access_token, refresh_token, expires_at, user_id, id_token, access_token_secret
	FROM providers
	WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var sp sqlProvider

	err := pm.DB.QueryRowContext(ctx, query, userId).Scan(
		&sp.Id,
		&sp.Provider,
		&sp.AccessToken,
		&sp.RefreshToken,
		&sp.ExpiresAt,
		&sp.UserId,
		&sp.IdToken,
		&sp.AccessTokenSecret,
	)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return validateSqlProvider(&sp), nil
}
