package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"rukenshia/frenchwhaling/pkg/storage"
	"rukenshia/frenchwhaling/pkg/wows"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/getsentry/sentry-go"
)

type Response = string

type Statistics []storage.EarnableResource

func Handler(ctx context.Context) (*Response, error) {
	data, err := storage.GetAllPublicSubscriberData()
	if err != nil {
		return nil, err
	}

	// The first is republic tokens which we don't care about :peeposhrug:
	resources := []*storage.EarnableResource{
		{Type: wows.Coal},
		{Type: wows.Steel},
		{Type: wows.SantaGiftContainer},
	}

	for _, subscriber := range data {
		for _, ship := range subscriber.Ships {
			for _, resourceType := range resources {
				if resourceType.Type == ship.Resource.Type {
					resourceType.Amount += ship.Resource.Amount
					resourceType.Earned += ship.Resource.Earned
				}
			}
		}
	}

	for _, resourceType := range resources {
		log.Printf("%+v", resourceType)
	}

	// Upload
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("eu-central-1"),
		},
	})
	if err != nil {
		return nil, err
	}
	svc := s3manager.NewUploader(sess)

	resourceData, err := json.Marshal(resources)
	if err != nil {
		return nil, err
	}

	if _, err := svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String("whaling.in.fkn.space"),
		Key:    aws.String("statistics.json"),
		Body:   bytes.NewBuffer(resourceData),
		ACL:    aws.String("public-read"),
	}); err != nil {
		return nil, err
	}

	res := Response("ok")
	return &res, nil
}

func main() {
	Handler(context.Background())
	sentry.Init(sentry.ClientOptions{
		Dsn:        os.Getenv("SENTRY_DSN"),
		ServerName: "generateGlobalStats",
	})

	lambda.Start(Handler)
}
