package main

import (
	"context"
	"log"
	"os"
	"rukenshia/frenchwhaling/pkg/auth"
	"rukenshia/frenchwhaling/pkg/storage"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/aws/aws-lambda-go/events"
	awsEvents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type E map[string]interface{}

func getHub(hub *sentry.Hub, fields map[string]interface{}) *sentry.Hub {
	h := hub.Clone()
	h.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetExtras(fields)
	})
	return h
}

// Handler is the lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsEvents.APIGatewayProxyRequest) (Response, error) {
	defer sentry.Flush(5 * time.Second)

	log.Printf("RequestRefresh start accountId=%s", request.PathParameters["accountId"])
	sentryAccountHub := sentry.CurrentHub().Clone()
	sentryAccountHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("AccountID", request.PathParameters["accountId"])
		scope.SetLevel(sentry.LevelError)
	})

	authz, ok := request.Headers["authorization"]
	if !ok {
		authz, ok = request.Headers["Authorization"]

		if !ok {
			return Response{
				StatusCode: 401,
				Body:       "No authorization passed",
				Headers: map[string]string{
					"Content-Type":                "text/plain",
					"Access-Control-Allow-Origin": "*",
				},
			}, nil
		}
	}

	if err := auth.VerifyToken(authz, request.PathParameters["accountId"]); err != nil {
		getHub(sentryAccountHub, E{"token": authz}).CaptureException(err)

		return Response{
			StatusCode: 401,
			Body:       "Unauthorized",
			Headers: map[string]string{
				"Content-Type":                "text/plain",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	log.Printf("Token verified, getting subscriber")

	subscriber, err := storage.GetSubscriber(request.PathParameters["accountId"])
	if err != nil {
		getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("GetSubscriber failed")
		log.Printf("ERROR: could not get subscriber accountId=%s error=%v", request.PathParameters["accountId"], err)

		return Response{
			StatusCode: 404,
			Body:       "Not found",
			Headers: map[string]string{
				"Content-Type":                "text/plain",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	log.Printf("LastScheduled accountId=%s scheduled=%d notBefore=%d", subscriber.AccountID, subscriber.LastScheduled, time.Now().Add(-1*time.Minute).UnixNano())

	if subscriber.LastScheduled > time.Now().Add(-10*time.Minute).UnixNano() {
		log.Printf("Preventing excessive updating, it has been done too recent accountId=%s", subscriber.AccountID)

		return Response{
			StatusCode: 400,
			Body:       "Too often",
			Headers: map[string]string{
				"Content-Type":                "text/plain",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	if err := subscriber.TriggerRefresh(); err != nil {
		getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("TriggerRefresh failed")
		log.Printf("ERROR: could not trigger refresh accountId=%s error=%v", subscriber.AccountID, err)

		return Response{
			StatusCode: 500,
			Body:       "Refresh failed",
			Headers: map[string]string{
				"Content-Type":                "text/plain",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	return Response{
		StatusCode: 200,
		Body:       "Started",
		Headers: map[string]string{
			"Content-Type":                "text/plain",
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn:        os.Getenv("SENTRY_DSN"),
		ServerName: "requestRefresh",
	})

	lambda.Start(Handler)
}
