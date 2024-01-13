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
	COMMENT_ADDED_EVENT     = "CommentAdded"
	SUB_COMMENT_ADDED_EVENT = "SubCommentAdded"
	COMMENT_BODY_MAX        = 100
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

func getBodyPreview(body string) string {
	if len(body) > COMMENT_BODY_MAX {
		return body[:COMMENT_BODY_MAX]
	} else {
		return body
	}
}

func (app *app) publishEvent(detail []byte) error {
	_, err := app.eb.PutEvents(&eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{
			{
				Detail:       aws.String(string(detail)),
				DetailType:   aws.String("NotificationReceived"),
				Source:       aws.String("notifications"),
				EventBusName: aws.String(os.Getenv("BUS_NAME")),
			},
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func (app *app) publishComment(comment *Comment) error {
	postUserId, err := app.models.Posts.GetPostUserId(comment.PostId)
	if err != nil {
		return err
	}

	if comment.User.Id == postUserId {
		return nil
	}

	event := CommentAddedEvent{
		PostId:              comment.PostId,
		CommentId:           comment.Id,
		PostUserId:          postUserId,
		CommentUserId:       comment.User.Id,
		CommentUserUsername: comment.User.Username,
		CommentCreatedAt:    comment.CreatedAt,
		CommentBodyPreview:  getBodyPreview(comment.Body),
		EventType:           COMMENT_ADDED_EVENT,
	}

	detail, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return app.publishEvent(detail)
}

type SubCommentAddedEvent struct {
	PostId                   int64     `json:"postId"`
	ParentCommentId          int64     `json:"parentCommentId"`
	ChildCommentId           int64     `json:"childCommentId"`
	ParentCommentUserId      int64     `json:"parentCommentUserId"`
	ChildCommentUserId       int64     `json:"childCommentUserId"`
	ChildCommentUserUsername string    `json:"childCommentUserUsername"`
	ChildCommentCreatedAt    time.Time `json:"childCommentCreatedAt"`
	ChildCommentBodyPreview  string    `json:"childCommentBody"`
	EventType                string    `json:"eventType"`
}

func (app *app) publishChildComment(comment *Comment) error {
	go func(comment *Comment) {
		err := app.publishComment(comment)
		if err != nil {
			fmt.Printf("Error publishing comment event: %s\n", err.Error())
		}
	}(comment)

	parentCommentUserId, err := app.models.Comments.getParentCommentUserId(comment.ParentId)
	if err != nil {
		return err
	}

	if comment.User.Id == parentCommentUserId {
		return nil
	}

	event := SubCommentAddedEvent{
		PostId:                   comment.PostId,
		ParentCommentId:          comment.ParentId,
		ChildCommentId:           comment.Id,
		ParentCommentUserId:      parentCommentUserId,
		ChildCommentUserId:       comment.User.Id,
		ChildCommentUserUsername: comment.User.Username,
		ChildCommentCreatedAt:    comment.CreatedAt,
		ChildCommentBodyPreview:  getBodyPreview(comment.Body),
		EventType:                SUB_COMMENT_ADDED_EVENT,
	}

	detail, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return app.publishEvent(detail)
}
