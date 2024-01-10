package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
)

const (
	POST_ADDED_EVENT = "PostAdded"
)

type EventBridge struct {
	eb *eventbridge.EventBridge
}

func NewEventBridge() *EventBridge {
	session := session.Must(session.NewSession())
	eb := eventbridge.New(session, aws.NewConfig().
		WithRegion("us-east-1").
		WithEndpoint("http://localstack:4566"),
	)

	return &EventBridge{
		eb: eb,
	}
}

type App struct {
	eb *EventBridge
}

type PostAddedEvent struct {
	PostId       string    `json:"postId"`
	UserId       int64     `json:"userId"`
	EventType    string    `json:"eventType"`
	Username     string    `json:"username"`
	SentAt       time.Time `json:"sentAt"`
	ConnectionId string    `json:"connectionId"`
}

// TO TEST PUBLISHING FOR NOW
func (app *App) handler(event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayV2HTTPResponse, error) {
	busName := os.Getenv("BUS_NAME")

	p := PostAddedEvent{
		PostId:       "123",
		UserId:       3,
		EventType:    POST_ADDED_EVENT,
		Username:     "bob-friend",
		SentAt:       time.Now(),
		ConnectionId: event.RequestContext.ConnectionID,
	}

	detail, err := json.Marshal(p)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			Body:       fmt.Sprintf("could not marshal event: %s", err.Error()),
			StatusCode: 500,
		}, nil
	}

	_, err = app.eb.eb.PutEvents(&eventbridge.PutEventsInput{
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
		return events.APIGatewayV2HTTPResponse{
			Body:       fmt.Sprintf("could not publish event: %s", err.Error()),
			StatusCode: 200,
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		Body:       fmt.Sprintf("event published"),
		StatusCode: 200,
	}, nil
}

func main() {
	eb := NewEventBridge()
	app := &App{eb: eb}
	lambda.Start(app.handler)
}
