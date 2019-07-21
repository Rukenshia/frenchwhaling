package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"rukenshia/frenchwhaling/pkg/storage"
	"rukenshia/frenchwhaling/pkg/wows/api"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dgrijalva/jwt-go"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is the lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	accessToken := request.QueryStringParameters["access_token"]
	accountId := request.QueryStringParameters["account_id"]
	realm := request.QueryStringParameters["realm"]
	// Verify the access token and account_id combination by making an authorized API call to the WG api
	res, err := api.GetPlayerInfo(realm, accessToken, accountId)
	if err != nil {
		log.Printf("Could not retrieve player info: %v", err)
		return Response{
			StatusCode: 302,
			Headers: map[string]string{
				"Location": "http://frenchwhaling.in.fkn.space/?success=false&reason=invalid-data",
			},
		}, nil
	}

	// Update DynDB
	subscriber, isNew, err := storage.FindOrCreateUpdateSubscriber(accessToken, realm, accountId)
	if err != nil {
		log.Printf("Could not crud subscriber info: %v", err)
		return Response{
			StatusCode: 302,
			Headers: map[string]string{
				"Location": fmt.Sprintf("http://frenchwhaling.in.fkn.space/?success=false&reason=subscription-failed&isNew=%t", isNew),
			},
		}, nil
	}

	log.Printf("Subscription=%v", subscriber)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":      "frenchwhaling",
		"exp":      request.QueryStringParameters["expires_at"],
		"nickname": res.Nickname,
		"sub":      res.AccountID,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SIGNING_SECRET")))
	if err != nil {
		log.Printf("Could not generate token: %v", err)
		return Response{
			StatusCode: 302,
			Headers: map[string]string{
				"Location": "http://frenchwhaling.in.fkn.space/?success=false&reason=signing-failed",
			},
		}, nil
	}

	resp := Response{
		StatusCode:      302,
		IsBase64Encoded: false,
		Body:            "",
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Location":     fmt.Sprintf("http://frenchwhaling.in.fkn.space/?success=true&isNew=%t&token=%s&dataUrl=%s", isNew, tokenString, subscriber.DataURL),
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
