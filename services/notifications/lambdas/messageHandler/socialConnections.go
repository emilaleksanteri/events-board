package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SocialConnsModel struct {
	DB *sql.DB
}

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type FriendNode struct {
	Id         int64
	UserId     int64
	nullId     sql.NullInt64
	nullUserId sql.NullInt64
}

func (s *SocialConnsModel) GetFriendsForUser(userId int64) ([]int64, error) {
	userNodeQuery := `
		select id, userid from friend_nodes where userid = $1
	`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var friendNode FriendNode
	err := s.DB.QueryRowContext(context, userNodeQuery, userId).Scan(
		&friendNode.Id,
		&friendNode.UserId,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	friendsQuery := `
		SELECT friend_nodes.id, friend_nodes.userId
		FROM friend_nodes
		JOIN friend_edges ON friend_nodes.id = friend_edges.next_node
		WHERE friend_edges.previous_node = $1;
	`

	rows, err := s.DB.QueryContext(context, friendsQuery, friendNode.Id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	var friends []int64
	for rows.Next() {
		var friend FriendNode
		err := rows.Scan(
			&friend.nullId,
			&friend.nullUserId,
		)

		if err != nil {
			return nil, err
		}

		if friend.nullUserId.Valid {
			friends = append(friends, friend.nullUserId.Int64)
		}
	}

	if rows.Err() != nil {
		return nil, err
	}

	return friends, nil
}
