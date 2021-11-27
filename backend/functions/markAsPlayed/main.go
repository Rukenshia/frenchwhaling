package main

import (
	"context"
	"log"
	"os"
	"rukenshia/frenchwhaling/pkg/auth"
	"rukenshia/frenchwhaling/pkg/events"
	"rukenshia/frenchwhaling/pkg/storage"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"

	awsEvents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response awsEvents.APIGatewayProxyResponse

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

	log.Printf("MarkAsPlayed start accountId=%s shipId=%s", request.PathParameters["accountId"], request.PathParameters["shipId"])
	sentryAccountHub := sentry.CurrentHub().Clone()
	sentryAccountHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("AccountID", request.PathParameters["accountId"])
		scope.SetTag("ShipID", request.PathParameters["shipId"])
		scope.SetLevel(sentry.LevelError)
	})

	authz, ok := request.Headers["authorization"]
	if !ok {
		authz, ok = request.Headers["Authorization"]

		if !ok {
			log.Printf("missing authz accountId=%s shipId=%s", request.PathParameters["accountId"], request.PathParameters["shipId"])
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
		log.Printf("token not valid err=%s accountId=%s shipId=%s", err.Error(), request.PathParameters["accountId"], request.PathParameters["shipId"])
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

	subscriberData, err := storage.LoadPublicSubscriberData(subscriber.DataURL)
	if err != nil {
		return Response{
			StatusCode: 500,
			Body:       "Could not find subscriber data",
			Headers: map[string]string{
				"Content-Type":                "text/plain",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	shipId, err := strconv.Atoi(request.PathParameters["shipId"])
	if err != nil {
		return Response{
			StatusCode: 400,
			Body:       "Bad ship id",
			Headers: map[string]string{
				"Content-Type":                "text/plain",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	shipId64 := int64(shipId)
	var ship *storage.StoredShip
	for _, knownShip := range subscriberData.Ships {
		if knownShip.ShipID == shipId64 {
			ship = knownShip
		}
	}

	if ship == nil {
		return Response{
			StatusCode: 400,
			Body:       "Unknown ship for player",
			Headers: map[string]string{
				"Content-Type":                "text/plain",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	if ship.Resource.Earned == ship.Resource.Amount {
		return Response{
			StatusCode: 400,
			Body:       "Already redeemed",
			Headers: map[string]string{
				"Content-Type":                "text/plain",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	ship.Resource.Earned = ship.Resource.Amount
	ship.ShipStatistics.LastBattleTime = int(time.Now().Unix())
	ship.LastBattleTime = int(time.Now().Unix())

	if err := events.Add(events.NewResourceEarned(subscriber.AccountID, ship.Resource.Type, ship.Resource.Amount, ship.ShipID, "manual")); err != nil {
		getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("Could not send ResourceEarned event")
		log.Printf("WARN: could not send resource earned event")
	}

	for i := range subscriberData.Resources {
		subscriberData.Resources[i].Earned = 0
	}

	for _, ship := range subscriberData.Ships {
		subscriberData.Resources[ship.Resource.Type].Earned += ship.Resource.Earned
	}

	if err := subscriberData.Save(subscriber.DataURL, false); err != nil {
		getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("Could not save data to S3")
		log.Printf("ERROR: Could not save data: accountId=%s error=%v", subscriber.AccountID, err)
		return Response{
			StatusCode: 500,
			Body:       "Error storing data",
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
