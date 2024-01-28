package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
)

const (
	POST_LIKE_EVENT = "PostLike"
)

type PostLikeEvent struct {
	PostId         int64     `json:"post_id"`
	PostUserId     int64     `json:"post_user_id"`
	PostLikeUserId int64     `json:"post_like_user_id"`
	EventType      string    `json:"event_type"`
	LikedAt        time.Time `json:"liked_at"`
}

func (app *app) publishPostLike(postId, postUserId, likeUserId int64) error {
	busName := os.Getenv("BUS_NAME")

	p := PostLikeEvent{
		PostId:         postId,
		PostUserId:     postUserId,
		PostLikeUserId: likeUserId,
		EventType:      POST_LIKE_EVENT,
		LikedAt:        time.Now(),
	}

	detail, err := json.Marshal(p)
	if err != nil {
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
		return err
	}

	return nil
}
