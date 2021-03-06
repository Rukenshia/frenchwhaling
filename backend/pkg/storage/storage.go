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
	AccountID            string
	Realm                string
	AccessToken          string
	AccessTokenExpiresAt int64
	DataURL              string
}

type Subscriber struct {
	Active               bool
	AccountID            string
	Realm                string
	AccessToken          string
	AccessTokenExpiresAt int64
	DataURL              string

	LastUpdated   int64
	LastScheduled int64
}

func GetSubscriber(accountId string) (*Subscriber, error) {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	log.Printf("GetSubscriber: start accountId=%s", accountId)
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("whaling-subscribers"),
		Key: map[string]*dynamodb.AttributeValue{
			"AccountID": {
				S: aws.String(accountId),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	item := Subscriber{}
	if err := dynamodbattribute.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal Record: %v", err)
	}

	return &item, nil
}

func FindOrCreateUpdateSubscriber(accessToken string, accessTokenExpiresAt int64, realm, accountId string) (*Subscriber, bool, error) {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	log.Printf("FindOrCreateUpdateSubscriber: start accountId=%s", accountId)
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("whaling-subscribers"),
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
			AccountID:            accountId,
			Realm:                realm,
			AccessToken:          accessToken,
			AccessTokenExpiresAt: accessTokenExpiresAt,
			DataURL:              getUniqueAccountURL(accountId),
			LastScheduled:        time.Now().UnixNano(),
			LastUpdated:          0,
			Active:               true,
		}

		av, err := dynamodbattribute.MarshalMap(subscriber)
		if err != nil {
			return nil, true, err
		}

		if _, err := svc.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String("whaling-subscribers"),
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
			TableName: aws.String("whaling-subscribers"),
			Item:      av,
		}); err != nil {
			return nil, false, err
		}
	}
	return &item, false, nil
}

func getUniqueAccountURL(accountID string) string {
	return fmt.Sprintf("https://whaling.in.fkn.space/data/%s/%s%s.json", accountID, xid.New().String(), xid.New().String())
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
	client := sns.New(session.New())

	r := []RefreshEvent{RefreshEvent{
		AccountID:            s.AccountID,
		Realm:                s.Realm,
		AccessToken:          s.AccessToken,
		AccessTokenExpiresAt: s.AccessTokenExpiresAt,
		DataURL:              s.DataURL,
	}}

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
				StringValue: aws.String("ManualRefresh"),
			},
		},
	})

	return err
}

// SetSubscriberActive sets the status of a subscriber to indicate whether they should be scheduled
//
// This is used mostly for when an access token expires prematurely or could not be refreshed. If
// the subscriber ends up logging in once more, they will be set to active again.
func SetSubscriberActive(accountID string, active bool) error {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	_, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("whaling-subscribers"),
		Key: map[string]*dynamodb.AttributeValue{
			"AccountID": {
				S: aws.String(accountID),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":a": {
				BOOL: aws.Bool(active),
			},
		},
		UpdateExpression: aws.String("set Active = :a"),
	})
	return err
}

func SetSubscriberAccessToken(accountID, accessToken string, expiresAt int64) error {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	_, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("whaling-subscribers"),
		Key: map[string]*dynamodb.AttributeValue{
			"AccountID": {
				S: aws.String(accountID),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":t": {
				S: aws.String(accessToken),
			},
			":e": {
				N: aws.String(fmt.Sprintf("%d", expiresAt)),
			},
		},
		UpdateExpression: aws.String("set AccessToken = :t, AccessTokenExpiresAt = :e"),
	})
	return err
}

func SetSubscriberLastUpdated(accountID string, timestamp int64) error {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	_, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("whaling-subscribers"),
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
		TableName: aws.String("whaling-subscribers"),
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

func getPage(lastEvaluated map[string]*dynamodb.AttributeValue, notScheduledSince int64) ([]*Subscriber, error) {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	out, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String("whaling-subscribers"),
		ExpressionAttributeNames: map[string]*string{
			"#ls": aws.String("LastScheduled"),
			"#a":  aws.String("Active"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":t": {
				N: aws.String(fmt.Sprintf("%d", notScheduledSince)),
			},
			":f": {
				BOOL: aws.Bool(true),
			},
		},
		FilterExpression:       aws.String("#ls < :t AND #a = :f"),
		ReturnConsumedCapacity: aws.String("TOTAL"),
		ExclusiveStartKey:      lastEvaluated,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("FindUnscheduledSubscribers: count=%d scanned=%d capacity=%f", *out.Count, *out.ScannedCount, *out.ConsumedCapacity.CapacityUnits)

	var subscribers []*Subscriber
	if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &subscribers); err != nil {
		return nil, err
	}

	if out.LastEvaluatedKey != nil {
		subs, err := getPage(out.LastEvaluatedKey, notScheduledSince)
		if err != nil {
			return nil, err
		}

		subscribers = append(subscribers, subs...)
	}

	return subscribers, nil
}

func FindUnscheduledSubscribers(notScheduledSince int64) ([]*Subscriber, error) {
	return getPage(nil, notScheduledSince)
}
