package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
)

const (
	COMMENT_ADDED_EVENT = "CommentAdded"
	COMMENT_BODY_MAX    = 100
)

type CommentAddedEvent struct {
	PostId              int64     `json:"postId"`
	CommentId           int64     `json:"commentId"`
	PostUserId          int64     `json:"postUserId"`
	CommentUserId       int64     `json:"commentUserId"`
	CommentUserUsername string    `json:"commentUserUsername"`
	CommentCreatedAt    time.Time `json:"commentCreatedAt"`
	CommentBodyPreview  string    `json:"commentBody"`
	EventType           string    `json:"eventType"`
}

// post event to bus on comment and sub comment
// -> event has to be sent to a connected post user id
// -> if sub comment, send to parent comment and post user id

func (app *app) publishComment(comment *Comment) error {
	busName := os.Getenv("BUS_NAME")

	post, err := app.models.Posts.GetPostWithUser(comment.PostId)
	if err != nil {
		return err
	}

	var commentBodyPreview string
	if len(comment.Body) > COMMENT_BODY_MAX {
		commentBodyPreview = comment.Body[:COMMENT_BODY_MAX]
	} else {
		commentBodyPreview = comment.Body
	}

	event := CommentAddedEvent{
		PostId:              comment.PostId,
		CommentId:           comment.Id,
		PostUserId:          post.UserId,
		CommentUserId:       comment.User.Id,
		CommentUserUsername: comment.User.Username,
		CommentCreatedAt:    comment.CreatedAt,
		CommentBodyPreview:  commentBodyPreview,
		EventType:           COMMENT_ADDED_EVENT,
	}

	detail, err := json.Marshal(event)
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
