package events

import (
	"rukenshia/frenchwhaling/pkg/wows"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type SubscriberEvent struct {
	AccountID string
	Timestamp int64
	Type      string
}

type ResourceEarned struct {
	SubscriberEvent
	ShipID     int64
	Resource   wows.Resource
	Amount     uint
	BattleType string
}

type ShipAddition struct {
	SubscriberEvent
	ShipID int64
}

func NewResourceEarned(accountID string, resource wows.Resource, amount uint, shipID int64, battleType string) ResourceEarned {
	return ResourceEarned{
		SubscriberEvent: SubscriberEvent{
			AccountID: accountID,
			Timestamp: time.Now().UnixNano(),
			Type:      "ResourceEarned",
		},
		ShipID:     shipID,
		Amount:     amount,
		Resource:   resource,
		BattleType: battleType,
	}
}

func NewShipAddition(accountID string, shipID int64) ShipAddition {
	return ShipAddition{
		SubscriberEvent: SubscriberEvent{
			AccountID: accountID,
			Timestamp: time.Now().UnixNano(),
			Type:      "ShipAddition",
		},
		ShipID: shipID,
	}
}

func Add(event interface{}) error {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(event)
	if err != nil {
		return err
	}

	if _, err := svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("frenchwhaling-subscriber-events"),
		Item:      av,
	}); err != nil {
		return err
	}

	return nil
}
