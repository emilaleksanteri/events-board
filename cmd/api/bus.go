package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/emilaleksanteri/pubsub/internal/data"
)

const (
	POST_ADDED    = "post-added"
	COMMENT_ADDED = "comment-added"
)

type EventPayload struct {
	Channel      string   `json:"channel"`
	Payload      string   `json:"payload"`
	Pattern      string   `json:"pattern"`
	PayloadSlice []string `json:"payloadSlice"`
}

type EventData struct {
	ID    string `json:"id"`
	Event string `json:"event"`
	Data  string `json:"data"`
	Retry int    `json:"retry"`
}

func (app *application) publishPostPostEvent(post *data.Post, ctx context.Context) error {
	payload, err := json.Marshal(post)
	if err != nil {
		return err
	}

	call := app.redis.Publish(ctx, POST_ADDED, payload)

	_, err = call.Result()
	if err != nil {
		return err
	}

	return nil
}

func (app *application) publishPostCommentEvent(comment *data.Comment, ctx context.Context) error {
	payload, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	call := app.redis.Publish(ctx, COMMENT_ADDED, payload)

	_, err = call.Result()
	if err != nil {
		return err
	}

	return nil
}

func (ed *EventData) String() string {
	ed.ID = fmt.Sprintf("%v", rand.Intn(100_000_000))

	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("id: %s\n", ed.ID))
	sb.WriteString(fmt.Sprintf("event: %s\n", ed.Event))
	sb.WriteString(fmt.Sprintf("data: %s\n", ed.Data))
	sb.WriteString(fmt.Sprintf("retry: %d\n\n", ed.Retry))

	return sb.String()
}

func (ed *EventData) Write(w io.Writer) (int64, error) {
	num, err := w.Write([]byte(ed.String()))
	if err != nil {
		return int64(num), err
	}

	return int64(num), nil
}

func ping(w io.Writer) error {
	payload := EventData{"", "ping", "", 1000}
	_, err := payload.Write(w)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) handleServerEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Println("SSE not supported buddy")
		http.Error(w, "SSE not supported buddy", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Transfer-Encoding", "chunked")

	err := ping(w)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-app.eventChan:
			event := EventData{"", msg.Channel, msg.Payload, 1000}

			event.Write(w)
			flusher.Flush()

		case <-time.After(time.Second * 30):
			err := ping(w)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			flusher.Flush()
		}
	}
}
