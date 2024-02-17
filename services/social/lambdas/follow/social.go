package main

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type FriendNode struct {
	Id   int64 `json:"id"`
	User User  `json:"user"`
}

type FiendEdge struct {
	PreviousNode FriendNode `json:"previous_node"`
	NextNode     FriendNode `json:"next_node"`
}

type SocialModel struct {
	DB *sql.DB
}

func (m *SocialModel) Follow(follower, following *User) error {
	query := `
		select id, userid from friend_nodes
		where userid = $1 or userid = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, follower.Id, following.Id)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var n FriendNode
		err = rows.Scan(&n.Id, &n.User.Id)
		if err != nil {
			return err
		}

		if n.User.Id == follower.Id {
			follower.Id = n.Id
		} else {
			following.Id = n.Id
		}
	}

	if err = rows.Err(); err != nil {
		return err
	}

	query = `
		insert into friend_edges
		(previous_node, next_node)
		values ($1, $2)
	`

	_, err = m.DB.ExecContext(ctx, query, follower.Id, following.Id)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint "):
			return ErrAlreadyFollowing
		default:
			return err
		}
	}

	return nil
}
