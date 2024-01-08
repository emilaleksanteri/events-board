package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
)

type EventBridge struct {
	eb *eventbridge.EventBridge
}

func NewEventBridge() *EventBridge {
	session := session.Must(session.NewSession())
	eb := eventbridge.New(session, aws.NewConfig().WithRegion("us-east-1").WithEndpoint("http://localstack:4566"))

	return &EventBridge{
		eb: eb,
	}
}

type App struct {
	eb *EventBridge
}

type EventBridgeEvent struct {
	ConnectionId   string `json:"connectionId"`
	NotificationId string `json:"notificationId"`
	Message        string `json:"message"`
}

func (app *App) handler(event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayV2HTTPResponse, error) {
	busName := os.Getenv("BUS_NAME")
	d := EventBridgeEvent{
		ConnectionId:   event.RequestContext.ConnectionID,
		NotificationId: "DEFAULT",
		Message:        "Hello from Lambda!",
	}

	detail, err := json.Marshal(d)
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
