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
	User       User      `json:"user"`
}

type CommentLike struct {
	Id         int64     `json:"id"`
	CommentId  int64     `json:"comment_id"`
	UserId     int64     `json:"user_id"`
	Created_at time.Time `json:"created_at"`
	User       User      `json:"user"`
}

type User struct {
	Id             int64  `json:"id"`
	ProfilePicture string `json:"profile_picture"`
	Username       string `json:"username"`
}

type Filter struct {
	take int
	skip int
}

type Metadata struct {
	TotalCount int `json:"total_count"`
	LeftCount  int `json:"left_count"`
}

func calculateMetadata(totalCount int, filter *Filter) Metadata {
	m := Metadata{}
	m.TotalCount = totalCount

	leftCount := 0
	if totalCount-(filter.skip+filter.take) > 0 {
		leftCount = totalCount - (filter.skip + filter.take)
	}

	m.LeftCount = leftCount

	return m
}

type PostLikesReturn struct {
	Likes    []PostLike `json:"likes"`
	Metadata Metadata   `json:"metadata"`
}

// TODO the get queries could be combined, at least some parts

func (p *LikeModel) getPostLikes(postId int64, filter *Filter) (*PostLikesReturn, error) {
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

	rows, err := p.DB.QueryContext(ctx, query, postId, filter.take, filter.skip)
	if err != nil {
		return nil, err
	}

	postLikes := []PostLike{}
	totalCount := 0
	defer rows.Close()
	for rows.Next() {
		var pl PostLike
		var u User
		err := rows.Scan(
			pl.Id,
			pl.PostId,
			pl.UserId,
			pl.Created_at,
			u.Id,
			u.Username,
			u.ProfilePicture,
			&totalCount,
		)

		if err != nil {
			return nil, err
		}

		pl.User = u
		postLikes = append(postLikes, pl)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	LikesReturn := PostLikesReturn{
		Likes:    postLikes,
		Metadata: calculateMetadata(totalCount, filter),
	}

	return &LikesReturn, nil
}

type CommentLikesReturn struct {
	Likes    []CommentLike `json:"likes"`
	Metadata Metadata      `json:"metadata"`
}

func (p *LikeModel) getCommentLikes(commentId int64, filter *Filter) (*CommentLikesReturn, error) {
	query := `
		select l.id, l.comment_id, l.user_id, l.created_at, 
		u.id, u.username, u.profile_picture,
		count(*) over() as full_count
		from comment_likes l
		join users u on u.id = l.user_id
		where l.comment_id = $1
		order by l.created_at desc
		limit $2 offset $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, query, commentId, filter.take, filter.skip)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	commentLikes := []CommentLike{}
	totalCount := 0
	for rows.Next() {
		var cl CommentLike
		var u User
		err := rows.Scan(
			cl.Id,
			cl.CommentId,
			cl.UserId,
			cl.Created_at,
			u.Id,
			u.Username,
			u.ProfilePicture,
			&totalCount,
		)

		if err != nil {
			return nil, err
		}

		cl.User = u
		commentLikes = append(commentLikes, cl)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	LikesReturn := CommentLikesReturn{
		Likes:    commentLikes,
		Metadata: calculateMetadata(totalCount, filter),
	}

	return &LikesReturn, nil
}
