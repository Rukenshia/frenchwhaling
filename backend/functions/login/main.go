package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"rukenshia/frenchwhaling/pkg/storage"
	"rukenshia/frenchwhaling/pkg/wows/api"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dgrijalva/jwt-go"
	"github.com/getsentry/sentry-go"
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
		scope.SetLevel(sentry.LevelWarning)
	})
	return h
}

// Handler is the lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	session := session.Must(session.NewSession())
	cloudwatchSvc := cloudwatch.New(session)

	defer sentry.Flush(5 * time.Second)
	accessToken := request.QueryStringParameters["access_token"]
	accountID := request.QueryStringParameters["account_id"]
	realm := request.QueryStringParameters["realm"]

	sentryAccountHub := sentry.CurrentHub().Clone()
	sentryAccountHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("AccountID", request.PathParameters["accountId"])
		scope.SetLevel(sentry.LevelError)
	})

	if param, ok := request.QueryStringParameters["message"]; ok {
		log.Printf("Auth has failed, aborting lambda accountId=%s reason=%s", accountID, strings.ToLower(param))

		cloudwatchSvc.PutMetricData(&cloudwatch.PutMetricDataInput{
			Namespace: aws.String("Whaling"),
			MetricData: []*cloudwatch.MetricDatum{
				{
					MetricName: aws.String("Login"),
					Dimensions: []*cloudwatch.Dimension{
						{Name: aws.String("Status"), Value: aws.String("Failed")},
						{Name: aws.String("Reason"), Value: aws.String(param)},
						{Name: aws.String("Realm"), Value: aws.String(realm)},
					},
					Value: aws.Float64(1.0),
				},
			},
		})

		return Response{
			StatusCode: 302,
			Headers: map[string]string{
				"Location": fmt.Sprintf("https://whaling.in.fkn.space/?success=false&reason=%s", strings.ToLower(param)),
			},
		}, nil
	}

	accessTokenExpiresAt, err := strconv.ParseInt(request.QueryStringParameters["expires_at"], 10, 64)
	if err != nil {
		getHub(sentryAccountHub, E{"query": request.QueryStringParameters, "error": err.Error()}).CaptureMessage("Could not parse expires_at")
		log.Printf("Could not parse expires_at: %v", err)
		return Response{
			StatusCode: 302,
			Headers: map[string]string{
				"Location": "https://whaling.in.fkn.space/?success=false&reason=invalid-expiry",
			},
		}, nil
	}

	// Verify the access token and account_id combination by making an authorized API call to the WG api
	res, err := api.GetPlayerInfo(realm, accessToken, accountID)
	if err != nil {
		getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("GetPlayerInfo failed")
		log.Printf("Could not retrieve player info: %v", err)
		return Response{
			StatusCode: 302,
			Headers: map[string]string{
				"Location": "https://whaling.in.fkn.space/?success=false&reason=invalid-data",
			},
		}, nil
	}

	// Update DynDB
	subscriber, isNew, err := storage.FindOrCreateUpdateSubscriber(accessToken, accessTokenExpiresAt, realm, accountID)
	if err != nil {
		getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("FindOrCreateUpdateSubscriber failed")
		log.Printf("Could not crud subscriber info: %v", err)
		return Response{
			StatusCode: 302,
			Headers: map[string]string{
				"Location": fmt.Sprintf("https://whaling.in.fkn.space/?success=false&reason=subscription-failed&isNew=%t", isNew),
			},
		}, nil
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":      "whaling",
		"exp":      request.QueryStringParameters["expires_at"],
		"nickname": res.Nickname,
		"realm":    request.QueryStringParameters["realm"],
		"sub":      fmt.Sprintf("%d", res.AccountID),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SIGNING_SECRET")))
	if err != nil {
		getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("Could not sign JWT")
		log.Printf("Could not generate token: %v", err)
		return Response{
			StatusCode: 302,
			Headers: map[string]string{
				"Location": "https://whaling.in.fkn.space/?success=false&reason=signing-failed",
			},
		}, nil
	}

	metricEvents := []*cloudwatch.MetricDatum{
		{
			MetricName: aws.String("Login"),
			Dimensions: []*cloudwatch.Dimension{
				{Name: aws.String("Status"), Value: aws.String("Success")},
				{Name: aws.String("Realm"), Value: aws.String(realm)},
			},
			Value: aws.Float64(1.0),
		},
	}

	if isNew == true {
		metricEvents = append(metricEvents, &cloudwatch.MetricDatum{
			MetricName: aws.String("FirstTimeLogin"),
			Dimensions: []*cloudwatch.Dimension{
				{Name: aws.String("Realm"), Value: aws.String(realm)},
			},
			Value: aws.Float64(1.0),
		})
	}

	if subscriber.Active == false {
		// The subscriber was previously deactivated, but now wants to use
		// whaling again. This is an interesting event, so lets put it to cloudwatch
		if err := storage.SetSubscriberActive(accountID, true); err != nil {
			getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("Could not re-enable subscriber")
		}

		metricEvents = append(metricEvents, &cloudwatch.MetricDatum{
			MetricName: aws.String("AccountReEnabled"),
			Dimensions: []*cloudwatch.Dimension{
				{Name: aws.String("Realm"), Value: aws.String(realm)},
			},
			Value: aws.Float64(1.0),
		})
	}

	cloudwatchSvc.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace:  aws.String("Whaling"),
		MetricData: metricEvents,
	})

	resp := Response{
		StatusCode:      302,
		IsBase64Encoded: false,
		Body:            "",
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Location":     fmt.Sprintf("https://whaling.in.fkn.space/?success=true&isNew=%t&token=%s&dataUrl=%s", isNew, tokenString, subscriber.DataURL),
		},
	}

	return resp, nil
}

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn:        os.Getenv("SENTRY_DSN"),
		ServerName: "login",
	})

	lambda.Start(Handler)
}
