package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"rukenshia/frenchwhaling/pkg/storage"
	"rukenshia/frenchwhaling/pkg/wows"
	"rukenshia/frenchwhaling/pkg/wows/api"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Handler is the lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event events.SNSEvent) (string, error) {
	var events []storage.RefreshEvent

	if err := json.Unmarshal([]byte(event.Records[0].SNS.Message), &events); err != nil {
		return "", fmt.Errorf("Could not parse event: %v", err)
	}

	// Process everything sequentially to avoid caring about rate limiting
	for _, ev := range events {
		log.Printf("Processing event: accountId=%s", ev.AccountID)

		log.Printf("Loading subscriber data: accountId=%s", ev.AccountID)

		var subscriberData *storage.SubscriberPublicData
		if sdata, err := storage.LoadPublicSubscriberData(ev.DataURL); err == nil {
			subscriberData = sdata
		} else {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == s3.ErrCodeNoSuchKey {
					log.Printf("Public data not found: will create new object later accountId=%s", ev.AccountID)

					now := time.Now()
					subscriberData = &storage.SubscriberPublicData{
						AccountID: ev.AccountID,
						Earnable: []storage.EarnabledResource{
							{Resource: wows.Coal, Amount: 0},
							{Resource: wows.RepublicTokens, Amount: 0},
						},
						Ships:       map[int64]*storage.StoredShip{},
						LastUpdated: &now,
					}
				} else {
					log.Printf("ERROR: Could not load subscriber data: accountId=%s code=%s error=%v", ev.AccountID, aerr.Code(), aerr)
					continue
				}
			} else {
				log.Printf("ERROR: Could not load subscriber data: accountId=%s error=%v", ev.AccountID, err)
				continue
			}
		}

		newData, err := api.GetPlayerShipStatistics(ev.AccessToken, ev.AccountID)
		if err != nil {
			log.Printf("ERROR: Processing event: failed for accountId=%s error=%v", ev.AccountID, err)
			continue
		}

		// Compare data
		log.Printf("Received data: comparing accountId=%s", ev.AccountID)

		for _, ship := range newData {
			wowsShip, ok := wows.Ships[ship.ShipID]

			if !wowsShip.IsEgligible() {
				continue
			}

			if !ok {
				log.Printf("ERROR: Ignoring unknown ship accountId=%s shipId=%d", ev.AccountID, ship.ShipID)
				continue
			}
			currentShip, ok := subscriberData.Ships[ship.ShipID]
			if !ok {
				// TODO: detect last battle time, set "Earned" automatically
				currentShip = &storage.StoredShip{
					ShipStatistics: ship,
					Earnable: storage.EarnabledResource{
						Resource: wowsShip.Resource(),
						Amount:   wowsShip.Amount(),
						Earned:   false,
					},
				}

				currentShip.LastBattleTime = 0
				currentShip.Pvp.Wins = 0

				log.Printf("New ship accountId=%s shipId=%d", ev.AccountID, ship.ShipID)
			}

			if ship.LastBattleTime > currentShip.LastBattleTime {
				// There is a new battle. Find out if it was a win
				win := false
				winType := ""

				if ship.Pvp.Wins > currentShip.Pvp.Wins {
					win = true
					winType = "pvp"
				} else if ship.Pve.Wins > currentShip.Pve.Wins {
					win = true
					winType = "pve"
				} else if ship.OperDiv.Wins > currentShip.OperDiv.Wins {
					win = true
					winType = "oper_div"
				} else if ship.OperSolo.Wins > currentShip.OperSolo.Wins {
					win = true
					winType = "oper_solo"
				} else if ship.RankSolo.Wins > currentShip.RankSolo.Wins {
					win = true
					winType = "rank_solo"
				}

				if win {
					log.Printf("Resource earned accountId=%s shipId=%d winType=%s", ev.AccountID, ship.ShipID, winType)

					// TODO: report win
					currentShip.ShipStatistics = ship
					currentShip.Earnable.Earned = true
					continue
				}
			}

			subscriberData.Ships[ship.ShipID] = currentShip
		}

		// Store data in S3
		if err := subscriberData.Save(ev.DataURL); err != nil {
			log.Printf("ERROR: Could not save data: accountId=%s error=%v", ev.AccountID, err)
			continue
		}
	}

	return fmt.Sprintf("Processed %d events", len(events)), nil
}

func main() {
	lambda.Start(Handler)
}
