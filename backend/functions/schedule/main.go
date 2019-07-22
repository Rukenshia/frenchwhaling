package main

import (
	"context"
	"log"
	"rukenshia/frenchwhaling/pkg/storage"
	"sync"
	"time"

	awsEvents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Handler is the lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsEvents.APIGatewayProxyRequest) (string, error) {
	log.Printf("Scheduler started")
	want := time.Now().Add(-30 * time.Minute)

	log.Printf("Finding last scheduled want=%d", want.UnixNano())

	subscribers, err := storage.FindUnscheduledSubscribers(want.UnixNano(), 600)
	if err != nil {
		log.Fatalf("Could not find subscribers: %v", err)
	}

	log.Printf("Found subscribers, sending refresh events subscribers=%d", len(subscribers))

	var batch []storage.RefreshEvent
	var wg sync.WaitGroup
	for _, subscriber := range subscribers {
		log.Printf("Selected for scheduling accountId=%s lastScheduled=%d", subscriber.AccountID, subscriber.LastScheduled)
		batch = append(batch, storage.RefreshEvent{
			AccountID:   subscriber.AccountID,
			Realm:       subscriber.Realm,
			AccessToken: subscriber.AccessToken,
			DataURL:     subscriber.DataURL,
		})

		wg.Add(1)
		go func(accountID string) {
			defer wg.Done()

			if err := storage.SetSubscriberLastScheduled(accountID, time.Now().UnixNano()); err != nil {
				log.Printf("ERROR: could not update last scheduled error=%v", err)
			}
		}(subscriber.AccountID)

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

	log.Printf("Waiting for subscriber schedule data update")
	wg.Wait()
	log.Printf("All subscriber scheduling info updated")

	return "done", nil
}

func main() {
	lambda.Start(Handler)
}
