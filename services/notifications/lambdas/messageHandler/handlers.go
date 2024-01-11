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
	POST_ADDED_EVENT = "PostAdded"
)

type PostAddedEvent struct {
	PostId    int64     `json:"postId"`
	UserId    int64     `json:"userId"`
	EventType string    `json:"eventType"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"sentAt"`
}

type Event struct {
	EventType string `json:"eventType"`
}

func (app *App) handler(event events.CloudWatchEvent) error {
	fmt.Printf("\n\nHANDLER HIT\n\n")
	var e Event

	err := json.Unmarshal(event.Detail, &e)
	if err != nil {
		fmt.Printf("Event detail missing EventType!\n")
		return err
	}

	fmt.Printf("Event: %v\n", e)

	var conns *[]NotificationRow
	switch e.EventType {
	case POST_ADDED_EVENT:
		var eventData PostAddedEvent
		err = json.Unmarshal(event.Detail, &eventData)
		if err != nil {
			fmt.Printf("Could not unmarshal event: %v\n", event.Detail)
			return err
		}

		conns, err = app.getConnections(eventData.UserId)
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
