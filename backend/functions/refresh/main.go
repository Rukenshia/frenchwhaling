package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"rukenshia/frenchwhaling/pkg/events"
	"rukenshia/frenchwhaling/pkg/storage"
	"rukenshia/frenchwhaling/pkg/wows"
	"rukenshia/frenchwhaling/pkg/wows/api"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	awsEvents "github.com/aws/aws-lambda-go/events"
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

// Handler is the lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event awsEvents.SNSEvent) (string, error) {
	defer sentry.Flush(5 * time.Second)
	var refreshEvents []storage.RefreshEvent

	if err := json.Unmarshal([]byte(event.Records[0].SNS.Message), &refreshEvents); err != nil {
		sentry.CaptureException(fmt.Errorf("Could not parse event: %v", err))
		return "", fmt.Errorf("Could not parse event: %v", err)
	}

	// Process everything sequentially to avoid caring about rate limiting
	for _, ev := range refreshEvents {
		sentryAccountHub := sentry.CurrentHub().Clone()
		sentryAccountHub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("AccountID", ev.AccountID)
		})

		if _, ok := wows.EventStartTime[ev.Realm]; !ok {
			log.Printf("WARN: Invalid realm for accountId=%s realm=%s", ev.AccountID, ev.Realm)
			sentryAccountHub.CaptureMessage(fmt.Sprintf("Invalid realm '%s'", ev.Realm))
			continue
		}

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
					getHub(sentryAccountHub, E{"error": aerr, "code": aerr.Code()}).CaptureMessage("Could not load subscriber data")
					log.Printf("ERROR: Could not load subscriber data: accountId=%s code=%s error=%v", ev.AccountID, aerr.Code(), aerr)
					continue
				}
			} else {
				getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("Could not load subscriber data")
				log.Printf("ERROR: Could not load subscriber data: accountId=%s error=%v", ev.AccountID, err)
				continue
			}
		}

		newData, err := api.GetPlayerShipStatistics(ev.Realm, ev.AccessToken, ev.AccountID)
		if err != nil {
			getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("GetPlayerShipStatistics failed")
			log.Printf("ERROR: Processing event: failed for accountId=%s error=%v", ev.AccountID, err)
			continue
		}

		// Get all ships in port
		shipsInPort, err := api.GetPlayerPort(ev.Realm, ev.AccessToken, ev.AccountID)
		if err != nil {
			getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("GetPlayerPort failed")
			log.Printf("ERROR: Could not retrieve ships in port accountId=%s error=%v", ev.AccountID, err)
			continue
		}

		// Add ships that were not in port before
		for _, shipID := range shipsInPort {
			wowsShip, ok := wows.Ships[shipID]
			if !ok {
				// Probably a ship that's not in the API anymore
				continue
			}

			if !wowsShip.IsEgligible() {
				continue
			}

			// If the data is not in subscriberData yet, we did not refresh it the last time
			if _, inCurrentData := subscriberData.Ships[shipID]; !inCurrentData {
				// We want to ignore ships that also have new statistics, it means the ship was already
				// played and will be processed further down.
				if _, inNewData := newData[shipID]; inNewData {
					continue
				}
			} else {
				continue
			}

			log.Printf("New ship found from port data accountId=%s shipId=%d", ev.AccountID, shipID)

			// Add the ship with empty data to newData,
			// this means it will be counted as ShipAddition further down
			newData[shipID] = api.ShipStatistics{
				ShipID:         shipID,
				LastBattleTime: -1,
				Private: api.ShipStatisticsPrivate{
					InGarage: true,
				},
			}
		}

		// Compare data
		log.Printf("Received data: comparing accountId=%s", ev.AccountID)

		for _, ship := range newData {
			sentryShipHub := sentryAccountHub.Clone()
			sentryShipHub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTag("ShipID", fmt.Sprintf("%d", ship.ShipID))
			})

			wowsShip, ok := wows.Ships[ship.ShipID]
			if !ok {
				// Probably a ship that doesn't really exist anymore
				continue
			}

			currentShip, ok := subscriberData.Ships[ship.ShipID]
			if !ok {
				if !wowsShip.IsEgligible() {
					continue
				}

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
					// send event
					if err := events.Add(events.NewShipAddition(ev.AccountID, ship.ShipID)); err != nil {
						getHub(sentryShipHub, E{"error": err.Error()}).CaptureMessage("Could not send ShipAddition event")
						log.Printf("WARN: could not send event for new subscriber ship error=%v", err)
					}
				}

				if ship.LastBattleTime > wows.EventStartTime[ev.Realm] {
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
							getHub(sentryShipHub, E{"error": err.Error()}).CaptureMessage("Could not send ResourceEarned event")
							log.Printf("WARN: could not send resource earned event")
						}
						continue
					}
				}
			}

			if !wowsShip.IsEgligible() {
				// remove the ship
				delete(subscriberData.Ships, ship.ShipID)
				log.Printf("Removed unegligible ship accountId=%s shipId=%d", ev.AccountID, ship.ShipID)

				continue
			}

			if currentShip.Resource.Earned > 0 {
				// Skip already earned ship
				currentShip.ShipStatistics = ship
				subscriberData.Ships[ship.ShipID] = currentShip
				continue
			}

			if ship.LastBattleTime != -1 && ship.LastBattleTime > currentShip.LastBattleTime && ship.LastBattleTime > wows.EventStartTime[ev.Realm] {
				// There is a new battle. Find out if it was a win and credit resources

				win, winType := getWinType(currentShip, ship)

				if win {
					currentShip.ShipStatistics = ship
					currentShip.Resource.Earned = currentShip.Resource.Amount

					if err := events.Add(events.NewResourceEarned(ev.AccountID, currentShip.Resource.Type, currentShip.Resource.Amount, currentShip.ShipID, winType)); err != nil {
						getHub(sentryShipHub, E{"error": err.Error()}).CaptureMessage("Could not send ResourceEarned event")
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
			getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("Could not save data to S3")
			log.Printf("ERROR: Could not save data: accountId=%s error=%v", ev.AccountID, err)
			continue
		}

		if err := storage.SetSubscriberLastUpdated(ev.AccountID, subscriberData.LastUpdated); err != nil {
			getHub(sentryAccountHub, E{"error": err.Error()}).CaptureMessage("Could not update LastUpdated in DynamoDB")
			log.Printf("ERROR: Could not set last updated accountId=%s error=%v", ev.AccountID, err)
		}
	}

	log.Printf("Processed all events count=%d", len(refreshEvents))

	return fmt.Sprintf("Processed %d refreshEvents", len(refreshEvents)), nil
}

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn:        os.Getenv("SENTRY_DSN"),
		ServerName: "refresh",
	})

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
