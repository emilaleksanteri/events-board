package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

const (
	POST_ADDED_EVENT    = "PostAdded"
	COMMENT_ADDED_EVENT = "CommentAdded"
)

type PostAddedEvent struct {
	PostId    int64     `json:"postId"`
	UserId    int64     `json:"userId"`
	EventType string    `json:"eventType"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"sentAt"`
}

type CommentAddedEvent struct {
	PostId             int64     `json:"postId"`
	CommentId          int64     `json:"commentId"`
	PostUserId         int64     `json:"postUserId"`
	CommentUserId      int64     `json:"commentUserId"`
	CommentUserName    string    `json:"commentUserName"`
	CommentCreatedAt   time.Time `json:"commentCreatedAt"`
	CommentBodyPreview string    `json:"commentBody"`
	EventType          string    `json:"eventType"`
}

type Event struct {
	EventType string `json:"eventType"`
}

func (app *App) handler(event events.CloudWatchEvent) error {
	var e Event

	err := json.Unmarshal(event.Detail, &e)
	if err != nil {
		fmt.Printf("Event detail missing EventType!\n")
		return err
	}

	var conns *[]NotificationRow
	switch e.EventType {
	case POST_ADDED_EVENT:
		var eventData PostAddedEvent
		err = json.Unmarshal(event.Detail, &eventData)
		if err != nil {
			fmt.Printf("Could not unmarshal event: %v\n", event.Detail)
			return err
		}

		conns, err = app.getConnectionsForPost(eventData.UserId)
		if err != nil {
			fmt.Printf("Could not get connections: %v\n", eventData)
			return err
		}

	case COMMENT_ADDED_EVENT:
		var eventData CommentAddedEvent
		err = json.Unmarshal(event.Detail, &eventData)
		if err != nil {
			fmt.Printf("Could not unmarshal event: %v\n", event.Detail)
			return err
		}

		conns, err = app.getPostAuthorConnection(eventData.PostUserId)
		if err != nil {
			fmt.Printf("Could not get connections: %v\n", eventData)
			return err
		}
	}

	for _, conn := range *conns {
		connection := conn
		go func(conn NotificationRow) {
			_, err = app.gw.PostToConnection(
				&apigatewaymanagementapi.PostToConnectionInput{
					ConnectionId: aws.String(conn.ConnectionId),
					Data:         event.Detail,
				})

			if err != nil {
				fmt.Printf("Could not send a msg to a conn: %s\n", err.Error())
			}
		}(connection)
	}

	return nil
}
