package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	awsEvents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is the lambda handler invoked by the `lambda.Start` function call
//
// This lambda function handles analytics events from the frontend and writes them
// straight to CloudWatch Metrics.
func Handler(ctx context.Context, request awsEvents.APIGatewayProxyRequest) (Response, error) {
	session := session.Must(session.NewSession())
	cloudwatchSvc := cloudwatch.New(session)

	validEvents := []string{
		"PrivacyPolicy", "Donate", "Contact", "Logout",
	}

	found := false
	for _, ev := range validEvents {
		if ev == request.Body {
			found = true
			break
		}
	}

	if !found {
		log.Printf("InvalidEventType type=%s", request.Body)
		return Response{
			StatusCode: 400,

			Headers: map[string]string{
				"Content-Type":                "text/plain",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	cloudwatchSvc.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String("Whaling"),
		MetricData: []*cloudwatch.MetricDatum{
			{
				MetricName: aws.String("ClickEvent"),
				Dimensions: []*cloudwatch.Dimension{
					{Name: aws.String("Type"), Value: aws.String(request.Body)},
				},
				Value: aws.Float64(1.0),
			},
		},
	})
	return Response{
		StatusCode: 200,

		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn:        os.Getenv("SENTRY_DSN"),
		ServerName: "click",
	})

	lambda.Start(Handler)
}
