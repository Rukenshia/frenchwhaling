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
	Realm       string
	AccessToken string
	DataURL     string
}

type Subscriber struct {
	AccountID   string
	Realm       string
	AccessToken string
	DataURL     string

	LastUpdated   int64
	LastScheduled int64
}

func FindOrCreateUpdateSubscriber(accessToken, realm, accountId string) (*Subscriber, bool, error) {
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
		return nil, false, err
	}

	if len(result.Item) == 0 {
		log.Printf("FindOrCreateUpdateSubscriber: creating new subscription accountId=%s", accountId)
		// Create entry
		subscriber := Subscriber{
			AccountID:     accountId,
			Realm:         realm,
			AccessToken:   accessToken,
			DataURL:       getUniqueAccountURL(accountId),
			LastScheduled: time.Now().UnixNano(),
			LastUpdated:   0,
		}

		av, err := dynamodbattribute.MarshalMap(subscriber)
		if err != nil {
			return nil, true, err
		}

		if _, err := svc.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String("frenchwhaling-subscribers"),
			Item:      av,
		}); err != nil {
			return nil, true, err
		}

		log.Printf("FindOrCreateUpdateSubscriber: trigger refresh accountId=%s", accountId)
		if err := subscriber.TriggerRefresh(); err != nil {
			return nil, true, err
		}

		return &subscriber, true, nil
	}

	item := Subscriber{}
	if err := dynamodbattribute.UnmarshalMap(result.Item, &item); err != nil {
		return nil, false, fmt.Errorf("Failed to unmarshal Record: %v", err)
	}

	// Update the access token
	if item.AccessToken != accessToken {
		log.Printf("FindOrCreateUpdateSubscriber: updating access token accountId=%s", accountId)
		item.AccessToken = accessToken

		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			return nil, false, err
		}

		if _, err := svc.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String("frenchwhaling-subscribers"),
			Item:      av,
		}); err != nil {
			return nil, false, err
		}
	}
	return &item, false, nil
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
		Realm:       s.Realm,
		AccessToken: s.AccessToken,
		DataURL:     s.DataURL,
	}

	return TriggerRefresh([]RefreshEvent{r})
}

func SetSubscriberLastUpdated(accountID string, timestamp int64) error {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	_, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("frenchwhaling-subscribers"),
		Key: map[string]*dynamodb.AttributeValue{
			"AccountID": {
				S: aws.String(accountID),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":l": {
				N: aws.String(fmt.Sprintf("%d", timestamp)),
			},
		},
		UpdateExpression: aws.String("set LastUpdated = :l"),
	})
	return err
}

func SetSubscriberLastScheduled(accountID string, timestamp int64) error {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	_, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("frenchwhaling-subscribers"),
		Key: map[string]*dynamodb.AttributeValue{
			"AccountID": {
				S: aws.String(accountID),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":l": {
				N: aws.String(fmt.Sprintf("%d", timestamp)),
			},
		},
		UpdateExpression: aws.String("set LastScheduled = :l"),
	})
	return err
}

func FindUnscheduledSubscribers(notScheduledSince, limit int64) ([]*Subscriber, error) {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	out, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String("frenchwhaling-subscribers"),
		ExpressionAttributeNames: map[string]*string{
			"#ls": aws.String("LastScheduled"),
		},
		Limit: &limit,
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":t": {
				N: aws.String(fmt.Sprintf("%d", notScheduledSince)),
			},
		},
		FilterExpression:       aws.String("#ls < :t"),
		ReturnConsumedCapacity: aws.String("TOTAL"),
	})
	if err != nil {
		return nil, err
	}
	log.Printf("FindUnscheduledSubscribers: count=%d scanned=%d capacity=%f", *out.Count, *out.ScannedCount, *out.ConsumedCapacity.CapacityUnits)

	var subscribers []*Subscriber
	if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &subscribers); err != nil {
		return nil, err
	}
	return subscribers, nil
}
