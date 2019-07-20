package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/rs/xid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type RefreshEvent struct {
	AccountID   string
	AccessToken string
	DataURL     string
}

type Subscriber struct {
	AccountID   string
	AccessToken string
	DataURL     string

	LastUpdated   *time.Time
	LastScheduled *time.Time
}

func FindOrCreateUpdateSubscriber(accessToken, accountId string) (*Subscriber, error) {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	log.Printf("FindOrCreateUpdateSubscriber: start accountId=%s", accountId)
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("frenchwhaling-subscribers"),
		Key: map[string]*dynamodb.AttributeValue{
			"AccountID": {
				S: aws.String(accountId),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(result.Item) == 0 {
		log.Printf("FindOrCreateUpdateSubscriber: creating new subscription accountId=%s", accountId)
		// Create entry
		subscriber := Subscriber{
			AccountID:     accountId,
			AccessToken:   accessToken,
			DataURL:       getUniqueAccountURL(accountId),
			LastUpdated:   nil,
			LastScheduled: nil,
		}

		av, err := dynamodbattribute.MarshalMap(subscriber)
		if err != nil {
			return nil, err
		}

		if _, err := svc.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String("frenchwhaling-subscribers"),
			Item:      av,
		}); err != nil {
			return nil, err
		}

		log.Printf("FindOrCreateUpdateSubscriber: trigger refresh accountId=%s", accountId)
		if err := subscriber.TriggerRefresh(); err != nil {
			return nil, err
		}

		return &subscriber, nil
	}

	item := Subscriber{}
	if err := dynamodbattribute.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal Record: %v", err)
	}

	// Update the access token
	if item.AccessToken != accessToken {
		log.Printf("FindOrCreateUpdateSubscriber: updating access token accountId=%s", accountId)
		item.AccessToken = accessToken

		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			return nil, err
		}

		if _, err := svc.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String("frenchwhaling-subscribers"),
			Item:      av,
		}); err != nil {
			return nil, err
		}
	}
	return &item, nil
}

func getUniqueAccountURL(accountID string) string {
	return fmt.Sprintf("https://frenchwhaling.in.fkn.space/data/%s/%s%s.json", accountID, xid.New().String(), xid.New().String())
}

func TriggerRefresh(r []RefreshEvent) error {
	client := sns.New(session.New())

	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	_, err = client.Publish(&sns.PublishInput{
		Message:  aws.String(string(data)),
		TopicArn: aws.String(os.Getenv("TOPIC_ARN")),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"Type": &sns.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String("Refresh"),
			},
		},
	})

	return err
}

func (s *Subscriber) TriggerRefresh() error {
	r := RefreshEvent{
		AccountID:   s.AccountID,
		AccessToken: s.AccessToken,
		DataURL:     s.DataURL,
	}

	return TriggerRefresh([]RefreshEvent{r})
}
