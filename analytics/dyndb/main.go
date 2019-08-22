package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"rukenshia/frenchwhaling/pkg/wows"

	"github.com/gocarina/gocsv"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func getPage(tableName *string, lastEvaluatedKey map[string]*dynamodb.AttributeValue) ([]map[string]interface{}, error) {
	session, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("eu-central-1"),
		},
	})
	if err != nil {
		return nil, err
	}

	svc := dynamodb.New(session)

	output, err := svc.Scan(&dynamodb.ScanInput{
		TableName:         tableName,
		ExclusiveStartKey: lastEvaluatedKey,
	})
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	if err := dynamodbattribute.UnmarshalListOfMaps(output.Items, &data); err != nil {
		return nil, err
	}

	if output.LastEvaluatedKey != nil {
		temp, err := getPage(tableName, output.LastEvaluatedKey)

		if err != nil {
			return nil, err
		}

		data = append(data, temp...)
	}

	return data, nil
}

type Event struct {
	AccountID string
	Timestamp uint64
	Type      string
	ShipID    int64

	Amount     *int64
	BattleType *string
	Resource   *int64
}

type EventStatistics struct {
	ShipName string
	Count    int64
	Sum      int64
}

type AccountStatistics struct {
	AccountID       string
	Battles         uint
	PveBattles      uint
	PvpBattles      uint
	OperSoloBattles uint
	OperDivBattles  uint
	RankSoloBattles uint
}

func getEventStats(eventType *string, events []*Event) []*EventStatistics {
	var out []*EventStatistics
	outMap := make(map[int64]*EventStatistics, 0)

	for _, e := range events {
		if e.Type != *eventType && *eventType != "" {
			continue
		}

		stats, ok := outMap[e.ShipID]
		if !ok {
			outMap[e.ShipID] = &EventStatistics{}
			stats = outMap[e.ShipID]

			stats.ShipName = wows.Ships[e.ShipID].Name
		}
		stats.Count++

		if e.Type == "ResourceEarned" {
			stats.Sum = stats.Sum + *e.Amount
		}
	}

	for _, s := range outMap {
		out = append(out, s)
	}

	return out
}

func main() {
	tableName := flag.String("table", "frenchwhaling-subscriber-events", "DynDB table name")
	refresh := flag.Bool("refresh", false, "Download data from table")
	eventType := flag.String("type", "ResourceEarned", "type (ResourceEarned, ShipAddition)")
	byAccount := flag.Bool("byaccount", false, "get results for each account")

	flag.Parse()

	if refresh != nil && *refresh {

		data, err := getPage(tableName, nil)
		if err != nil {
			log.Fatal(err)
		}

		buf, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}

		if err := ioutil.WriteFile(path.Join("data", "events.json"), buf, 0666); err != nil {
			log.Fatal(err)
		}
	}

	// load file again
	data, err := ioutil.ReadFile(path.Join("data", "events.json"))
	if err != nil {
		log.Fatal(err)
	}

	var events []*Event
	if err := json.Unmarshal(data, &events); err != nil {
		log.Fatal(err)
	}

	if *byAccount {
		accountMap := make(map[string]*AccountStatistics)

		for _, e := range events {
			if e.Type != "ResourceEarned" {
				continue
			}

			a, ok := accountMap[e.AccountID]
			if !ok {
				accountMap[e.AccountID] = &AccountStatistics{}
				a = accountMap[e.AccountID]
			}

			a.AccountID = e.AccountID

			a.Battles++

			switch *e.BattleType {
			case "oper_solo":
				a.OperSoloBattles++
			case "oper_div":
				a.OperDivBattles++
			case "pve":
				a.PveBattles++
			case "pvp":
				a.PvpBattles++
			case "rank_solo":
				a.RankSoloBattles++
			}
		}

		var out []*AccountStatistics

		for _, s := range accountMap {
			out = append(out, s)
		}

		err = gocsv.Marshal(out, os.Stdout)
		if err != nil {
			panic(err)
		}

	} else {
		out := getEventStats(eventType, events)

		err = gocsv.Marshal(out, os.Stdout)
		if err != nil {
			panic(err)
		}
	}
}
