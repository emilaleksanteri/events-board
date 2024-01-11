package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
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

func (app *app) publishPost(post *Post) error {
	busName := os.Getenv("BUS_NAME")

	p := PostAddedEvent{
		PostId:    post.Id,
		UserId:    post.User.Id,
		EventType: POST_ADDED_EVENT,
		Username:  post.User.Username,
		CreatedAt: post.CreatedAt,
	}

	detail, err := json.Marshal(p)
	if err != nil {
		fmt.Printf("Invalid detail schema: %v\n", err)
		return err
	}

	_, err = app.eb.PutEvents(&eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{
			{
				Detail:       aws.String(string(detail)),
				DetailType:   aws.String("NotificationReceived"),
				Source:       aws.String("notifications"),
				EventBusName: aws.String(busName),
			},
		},
	})

	if err != nil {
		fmt.Printf("Could not publish event for post %d: \n%v\n", post.Id, err)
		return err
	}

	return nil
}
