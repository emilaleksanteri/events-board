package main

import (
	"context"
	"database/sql"
	"time"
)

type LikeModel struct {
	DB *sql.DB
}

type PostLike struct {
	Id         int64     `json:"id"`
	PostId     int64     `json:"post_id"`
	UserId     int64     `json:"user_id"`
	Created_at time.Time `json:"created_at"`
	User       *User     `json:"user"`
}

type CommentLike struct {
	Id         int64     `json:"id"`
	CommentId  int64     `json:"comment_id"`
	UserId     int64     `json:"user_id"`
	Created_at time.Time `json:"created_at"`
	User       *User     `json:"user"`
}

type User struct {
	Id                int64          `json:"id"`
	ProfilePicture    string         `json:"profile_picture"`
	Username          string         `json:"username"`
	sqlID             sql.NullInt64  `json:"-"`
	sqlProfilePicture sql.NullString `json:"-"`
	sqlUsername       sql.NullString `json:"-"`
}

func (u *User) parseSqlNulls() {
	if u.sqlID.Valid {
		u.Id = u.sqlID.Int64
	}

	if u.sqlProfilePicture.Valid {
		u.ProfilePicture = u.sqlProfilePicture.String
	}

	if u.sqlUsername.Valid {
		u.Username = u.sqlUsername.String
	}
}

type Filter struct {
	take int
	skip int
}

type Metadata struct {
	TotalCount int `json:"total_count"`
	LeftCount  int `json:"left_count"`
}

type LikesReturn struct {
	Likes    *[]PostLike `json:"likes"`
	Metadata *Metadata   `json:"metadata"`
}

func (p *LikeModel) getLikes(postId int64, filter *Filter) (*LikesReturn, error) {
	query := `
		select l.id, l.post_id, l.user_id, l.created_at, 
		u.id, u.username, u.profile_picture,
		count(*) over() as full_count
		from post_likes l
		join users u on u.id = l.user_id
		where l.post_id = $1
		order by l.created_at desc
		limit $2 offset $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{postId, filter.take, filter.skip}
	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	postLikes := []PostLike{}
	totalCount := 0
	for rows.Next() {
		var pl PostLike
		var u User
		err := rows.Scan(
			&pl.Id,
			&pl.PostId,
			&pl.UserId,
			&pl.Created_at,
			&u.sqlID,
			&u.sqlUsername,
			&u.sqlProfilePicture,
			&totalCount,
		)

		if err != nil {
			return nil, err
		}

		u.parseSqlNulls()
		if u.Id != 0 {
			pl.User = &u
		}

		postLikes = append(postLikes, pl)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	metadata := Metadata{
		TotalCount: totalCount,
		LeftCount:  totalCount - (filter.skip + filter.take),
	}

	LikesReturn := LikesReturn{
		Likes:    &postLikes,
		Metadata: &metadata,
	}

	return &LikesReturn, nil
}
