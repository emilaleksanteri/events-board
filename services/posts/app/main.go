package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch event.Path {
	case "/":
		response := events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "\"hello from lambda!\"",
		}
		return response, nil
	case "/health":
		response := events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "\"running healthy :)\"",
		}

		return response, nil
	default:
		response := events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "\"not found\"",
		}
		return response, nil
	}
}

func main() {
	lambda.Start(handler)
}
