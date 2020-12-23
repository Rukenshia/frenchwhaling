package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"rukenshia/frenchwhaling/pkg/storage"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/aws/aws-lambda-go/lambda"
)

type E map[string]interface{}

func getHub(hub *sentry.Hub, fields map[string]interface{}) *sentry.Hub {
	h := hub.Clone()
	h.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetExtras(fields)
	})
	return h
}

// Request is the payload the service gets called with
type Request struct {
	RefreshAll bool
}

// Handler is the lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request Request) (string, error) {
	defer sentry.Flush(5 * time.Second)
	log.Printf("Scheduler started")
	want := time.Now().Add(-120 * time.Minute)

	if request.RefreshAll {
		want = time.Now()
	}

	log.Printf("Finding last scheduled want=%d", want.UnixNano())

	subscribers, err := storage.FindUnscheduledSubscribers(want.UnixNano())
	if err != nil {
		getHub(sentry.CurrentHub(), E{"error": err}).CaptureException(fmt.Errorf("FindUnscheduledSubscribers failed"))
		log.Fatalf("Could not find subscribers: %v", err)
	}

	log.Printf("Found subscribers, sending refresh events subscribers=%d", len(subscribers))

	var batch []storage.RefreshEvent
	var wg sync.WaitGroup
	for _, subscriber := range subscribers {
		sentryAccountHub := sentry.CurrentHub().Clone()
		sentryAccountHub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("AccountID", subscriber.AccountID)
		})

		log.Printf("Selected for scheduling accountId=%s lastScheduled=%d", subscriber.AccountID, subscriber.LastScheduled)
		batch = append(batch, storage.RefreshEvent{
			AccountID:            subscriber.AccountID,
			Realm:                subscriber.Realm,
			AccessToken:          subscriber.AccessToken,
			AccessTokenExpiresAt: subscriber.AccessTokenExpiresAt,
			DataURL:              subscriber.DataURL,
		})

		wg.Add(1)
		go func(accountID string) {
			defer wg.Done()

			if err := storage.SetSubscriberLastScheduled(accountID, time.Now().UnixNano()); err != nil {
				sentryAccountHub.CaptureException(fmt.Errorf("SetSubscriberLastScheduled failed"))
				log.Printf("ERROR: could not update last scheduled error=%v", err)
			}
		}(subscriber.AccountID)

		if len(batch) >= 100 {
			log.Printf("Sending batch of size=%d", len(batch))

			if err := storage.TriggerRefresh(batch); err != nil {
				sentry.CaptureException(fmt.Errorf("TriggerRefresh failed"))
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
		sentry.CaptureException(fmt.Errorf("TriggerRefresh failed"))
		log.Printf("ERROR: sending batch error=%v", err)
	}

	log.Printf("Waiting for subscriber schedule data update")
	wg.Wait()
	log.Printf("All subscriber scheduling info updated")

	return "done", nil
}

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn:        os.Getenv("SENTRY_DSN"),
		ServerName: "schedule",
	})

	lambda.Start(Handler)
}
