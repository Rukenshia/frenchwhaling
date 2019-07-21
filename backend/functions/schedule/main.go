package main

import (
	"context"
	"log"
	"rukenshia/frenchwhaling/pkg/storage"
	"time"

	awsEvents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Handler is the lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsEvents.APIGatewayProxyRequest) (string, error) {
	log.Printf("Scheduler started")
	want := time.Now().Add(-15 * time.Minute)

	subscribers, err := storage.FindUnscheduledSubscribers(want.UnixNano(), 600)
	if err != nil {
		log.Fatalf("Could not find subscribers: %v", err)
	}

	log.Printf("Found subscribers, sending refresh events subscribers=%d", len(subscribers))

	var batch []storage.RefreshEvent
	for _, subscriber := range subscribers {
		batch = append(batch, storage.RefreshEvent{
			AccountID:   subscriber.AccountID,
			Realm:       subscriber.Realm,
			AccessToken: subscriber.AccessToken,
			DataURL:     subscriber.DataURL,
		})

		go func() {
			if err := storage.SetSubscriberLastScheduled(subscriber.AccountID, time.Now().UnixNano()); err != nil {
				log.Printf("ERROR: could not update last scheduled error=%v", err)
			}
		}()

		if len(batch) >= 100 {
			log.Printf("Sending batch of size=%d", len(batch))

			if err := storage.TriggerRefresh(batch); err != nil {
				log.Printf("ERROR: sending batch error=%v", err)
			}
			batch = []storage.RefreshEvent{}
		}
	}

	if len(batch) == 0 {
		log.Printf("Skipping last batch, no items")
		return "done", nil
	}

	// send last batch
	log.Printf("Sending last batch of size=%d", len(batch))

	if err := storage.TriggerRefresh(batch); err != nil {
		log.Printf("ERROR: sending batch error=%v", err)
	}

	return "done", nil
}

func main() {
	lambda.Start(Handler)
}
