package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/emilaleksanteri/pubsub/internal/data"
)

const (
	POST_ADDED    = "post-added"
	COMMENT_ADDED = "comment-added"
)

func (app *application) publishPostPostEvent(post *data.Post) error {
	payload, err := json.Marshal(post)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	call := app.redis.Publish(ctx, POST_ADDED, payload)

	_, err = call.Result()
	if err != nil {
		return err
	}

	return nil
}

func (app *application) publishPostCommentEvent(comment *data.Comment) error {
	payload, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	call := app.redis.Publish(ctx, COMMENT_ADDED, payload)

	_, err = call.Result()
	if err != nil {
		return err
	}

	return nil
}

func formatServerEvent(event string, data string) (string, error) {
	stringBuilder := strings.Builder{}

	stringBuilder.WriteString(fmt.Sprintf("event: %s\n", event))
	stringBuilder.WriteString(fmt.Sprintf("data: %v\n\n", data))

	return stringBuilder.String(), nil

}

func (app *application) handleServerEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported buddy", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Transfer-Encoding", "chunked")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sub := app.redis.Subscribe(ctx, POST_ADDED, COMMENT_ADDED)
	defer sub.Close()

	channel := sub.Channel()

	for msg := range channel {

		eventData, err := formatServerEvent(msg.Channel, msg.Payload)
		if err != nil {
			fmt.Println(err)
			sub.Unsubscribe(ctx, POST_ADDED, COMMENT_ADDED)
			app.serverErrorResponse(w, r, err)
			break
		}

		w.Write([]byte(eventData))
		if err != nil {
			fmt.Println(err)
			sub.Unsubscribe(ctx, POST_ADDED, COMMENT_ADDED)
			app.serverErrorResponse(w, r, err)
			break
		}

		flusher.Flush()
	}
}
