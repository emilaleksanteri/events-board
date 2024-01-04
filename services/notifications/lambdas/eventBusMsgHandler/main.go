package main

import (
	"encoding/json"
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
	eb := eventbridge.New(session)

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

func (app *App) handler(event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	busName := os.Getenv("BUS_NAME")
	d := EventBridgeEvent{
		ConnectionId:   event.RequestContext.ConnectionID,
		NotificationId: "DEFAULT",
		Message:        "Hello from Lambda!",
	}

	detail, err := json.Marshal(d)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, nil
	}

	entry := eventbridge.PutEventsRequestEntry{
		Detail:       aws.String(string(detail)),
		DetailType:   aws.String("notification"),
		EventBusName: aws.String(busName),
		Source:       aws.String("notification"),
	}

	_, err = app.eb.eb.PutEvents(&eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{
			&entry,
		},
	})

	return events.APIGatewayProxyResponse{
		Body:       "OK",
		StatusCode: 200,
	}, nil

}

func main() {
	eb := NewEventBridge()
	app := &App{eb: eb}
	lambda.Start(app.handler)
}
