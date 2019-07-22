package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"rukenshia/frenchwhaling/pkg/events"
	"rukenshia/frenchwhaling/pkg/storage"
	"rukenshia/frenchwhaling/pkg/wows"
	"rukenshia/frenchwhaling/pkg/wows/api"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	awsEvents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var EventStartTime = map[string]int{
	"eu":   1563710005,
	"com":  1563710005,
	"ru":   1563710005,
	"asia": 1563710005,
}

// Handler is the lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event awsEvents.SNSEvent) (string, error) {
	var refreshEvents []storage.RefreshEvent

	if err := json.Unmarshal([]byte(event.Records[0].SNS.Message), &refreshEvents); err != nil {
		return "", fmt.Errorf("Could not parse event: %v", err)
	}

	// Process everything sequentially to avoid caring about rate limiting
	for _, ev := range refreshEvents {
		log.Printf("Loading subscriber data: accountId=%s", ev.AccountID)

		isNewSubscriber := false
		var subscriberData *storage.SubscriberPublicData
		if sdata, err := storage.LoadPublicSubscriberData(ev.DataURL); err == nil {
			subscriberData = sdata
		} else {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == s3.ErrCodeNoSuchKey {
					log.Printf("Public data not found: will create new object later accountId=%s", ev.AccountID)

					subscriberData = &storage.SubscriberPublicData{
						AccountID: ev.AccountID,
						Resources: []storage.EarnableResource{
							{Type: wows.RepublicTokens, Amount: 0, Earned: 0},
							{Type: wows.Coal, Amount: 0, Earned: 0},
						},
						Ships:       map[int64]*storage.StoredShip{},
						LastUpdated: time.Now().UnixNano(),
					}
					isNewSubscriber = true
				} else {
					log.Printf("ERROR: Could not load subscriber data: accountId=%s code=%s error=%v", ev.AccountID, aerr.Code(), aerr)
					continue
				}
			} else {
				log.Printf("ERROR: Could not load subscriber data: accountId=%s error=%v", ev.AccountID, err)
				continue
			}
		}

		newData, err := api.GetPlayerShipStatistics(ev.Realm, ev.AccessToken, ev.AccountID)
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
					Resource: storage.EarnableResource{
						Type:   wowsShip.Resource(),
						Amount: wowsShip.Amount(),
						Earned: 0,
					},
				}

				if !isNewSubscriber {
					// Reset all battles on the ship

					// send event
					if err := events.Add(events.NewShipAddition(ev.AccountID, ship.ShipID)); err != nil {
						log.Printf("WARN: could not send event for new subscriber ship error=%v", err)
					}
				}

				if ship.LastBattleTime > EventStartTime["eu"] {
					// A battle was played with a ship that we did not know yet.
					// For new subscribers, they might be coming to the event late.
					// For existing subscribers, they might just have bought a ship and played a battle
					// with it. Let's give them the resource if we can find any wins.

					// Compare against empty statistics to find a win
					win, winType := getWinType(&storage.StoredShip{
						ShipStatistics: api.ShipStatistics{},
					}, ship)

					if win {
						// Credit the resources
						currentShip.ShipStatistics = ship
						currentShip.Resource.Earned = currentShip.Resource.Amount
						subscriberData.Ships[ship.ShipID] = currentShip

						if err := events.Add(events.NewResourceEarned(ev.AccountID, currentShip.Resource.Type, currentShip.Resource.Amount, currentShip.ShipID, winType)); err != nil {
							log.Printf("WARN: could not send resource earned event")
						}
						continue
					}
				}
			}

			if ship.LastBattleTime > currentShip.LastBattleTime {
				// There is a new battle. Find out if it was a win and credit resources

				win, winType := getWinType(currentShip, ship)

				if win {
					currentShip.ShipStatistics = ship
					currentShip.Resource.Earned = currentShip.Resource.Amount

					if err := events.Add(events.NewResourceEarned(ev.AccountID, currentShip.Resource.Type, currentShip.Resource.Amount, currentShip.ShipID, winType)); err != nil {
						log.Printf("WARN: could not send resource earned event")
					}
				}
			}

			currentShip.ShipStatistics = ship
			subscriberData.Ships[ship.ShipID] = currentShip
		}

		for i := range subscriberData.Resources {
			subscriberData.Resources[i].Earned = 0
		}

		for _, ship := range subscriberData.Ships {
			subscriberData.Resources[ship.Resource.Type].Earned += ship.Resource.Earned
		}

		subscriberData.LastUpdated = time.Now().UnixNano()

		// Store data in S3
		if err := subscriberData.Save(ev.DataURL); err != nil {
			log.Printf("ERROR: Could not save data: accountId=%s error=%v", ev.AccountID, err)
			continue
		}

		if err := storage.SetSubscriberLastUpdated(ev.AccountID, subscriberData.LastUpdated); err != nil {
			log.Printf("ERROR: Could not set last updated accountId=%s error=%v", ev.AccountID, err)
		}
	}

	log.Printf("Processed all events count=%d", len(refreshEvents))

	return fmt.Sprintf("Processed %d refreshEvents", len(refreshEvents)), nil
}

func main() {
	lambda.Start(Handler)
}

func getWinType(currentShip *storage.StoredShip, newShip api.ShipStatistics) (win bool, winType string) {
	if newShip.Pvp.Wins > currentShip.Pvp.Wins {
		win = true
		winType = "pvp"
	} else if newShip.Pve.Wins > currentShip.Pve.Wins {
		win = true
		winType = "pve"
	} else if newShip.OperDiv.Wins > currentShip.OperDiv.Wins {
		win = true
		winType = "oper_div"
	} else if newShip.OperSolo.Wins > currentShip.OperSolo.Wins {
		win = true
		winType = "oper_solo"
	} else if newShip.RankSolo.Wins > currentShip.RankSolo.Wins {
		win = true
		winType = "rank_solo"
	}
	return win, winType
}
