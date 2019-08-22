package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"rukenshia/frenchwhaling/pkg/storage"
	"rukenshia/frenchwhaling/pkg/wows"
	"sort"

	"github.com/gocarina/gocsv"
)

type MergedFile struct {
	First *storage.SubscriberPublicData `json:"first"`
	Last  *storage.SubscriberPublicData `json:"last"`
}

type Statistics struct {
	AccountID              string
	CoalEarned             uint
	CoalEarnable           uint
	TokensEarned           uint
	TokensEarnable         uint
	ShipsInPortStart       uint
	ShipsInPortEnd         uint
	BattlesPlayed          int
	BattlesWon             int
	BattlesPvp             int
	BattlesPvpWon          int
	BattlesRankSolo        int
	BattlesRankSoloWon     int
	BattlesOperDiv         int
	BattlesOperDivWon      int
	BattlesPve             int
	BattlesPveWon          int
	BattlesOperSolo        int
	BattlesOperSoloWon     int
	HasJaguar              bool
	HasGuepard             bool
	HasVauquelin           bool
	HasLeFantasque         bool
	HasMogador             bool
	MostPlayedShip1        string `csv:"-"`
	MostPlayedShip1Battles int    `csv:"-"`
	MostPlayedShip2        string `csv:"-"`
	MostPlayedShip2Battles int    `csv:"-"`
	MostPlayedShip3        string `csv:"-"`
	MostPlayedShip3Battles int    `csv:"-"`
}

var (
	Mogador     = int64(4180555600)
	LeFantasque = int64(4181604176)
	Vauquelin   = int64(4182652752)
	Guépard     = int64(4183701328)
	Jaguar      = int64(4184749904)
)

func main() {
	in := flag.String("in", "", "input file")

	flag.Parse()

	file, err := readFile(*in)
	if err != nil {
		log.Fatal(err)
	}

	s := Statistics{}

	s.AccountID = file.First.AccountID
	s.CoalEarned = file.Last.Resources[wows.Coal].Earned
	s.TokensEarned = file.Last.Resources[wows.RepublicTokens].Earned

	for idx, ship := range file.Last.Ships {
		firstShip := file.First.Ships[idx]

		if firstShip == nil {
			firstShip = &storage.StoredShip{}
		}

		s.BattlesPlayed += ship.Battles - firstShip.Battles

		s.BattlesWon += (ship.Pvp.Wins + ship.RankSolo.Wins + ship.OperDiv.Wins + ship.Pve.Wins + ship.OperSolo.Wins) - (firstShip.Pvp.Wins + firstShip.RankSolo.Wins + firstShip.OperDiv.Wins + firstShip.Pve.Wins + firstShip.OperSolo.Wins)

		if firstShip.Private.InGarage {
			s.ShipsInPortStart++
		}

		s.BattlesPvp += ship.Pvp.Battles - firstShip.Pvp.Battles
		s.BattlesPvpWon += ship.Pvp.Wins - firstShip.Pvp.Wins
		s.BattlesRankSolo += ship.RankSolo.Battles - firstShip.RankSolo.Battles
		s.BattlesRankSoloWon += ship.RankSolo.Wins - firstShip.RankSolo.Wins
		s.BattlesOperDiv += ship.OperDiv.Battles - firstShip.OperDiv.Battles
		s.BattlesOperDivWon += ship.OperDiv.Wins - firstShip.OperDiv.Wins
		s.BattlesPve += ship.Pve.Battles - firstShip.Pve.Battles
		s.BattlesPveWon += ship.Pve.Wins - firstShip.Pve.Wins
		s.BattlesOperSolo += ship.OperSolo.Battles - firstShip.OperSolo.Battles
		s.BattlesOperSoloWon += ship.OperSolo.Wins - firstShip.OperSolo.Wins

		if ship.Private.InGarage {
			s.ShipsInPortEnd++
		}

		if ship.Resource.Type == wows.Coal {
			s.CoalEarnable += ship.Resource.Amount
		} else if ship.Resource.Type == wows.RepublicTokens {
			s.TokensEarnable += ship.Resource.Amount
		}

		switch ship.ShipID {
		case Mogador:
			s.HasMogador = true
		case LeFantasque:
			s.HasLeFantasque = true
		case Vauquelin:
			s.HasVauquelin = true
		case Guépard:
			s.HasGuepard = true
		case Jaguar:
			s.HasJaguar = true
		default:
			continue

		}
	}

	var shipSlice []*storage.StoredShip
	for _, ship := range file.Last.Ships {
		shipSlice = append(shipSlice, ship)
	}

	sort.SliceStable(shipSlice, func(i, j int) bool {
		firstShipI := file.First.Ships[shipSlice[i].ShipID]
		firstShipJ := file.First.Ships[shipSlice[j].ShipID]

		ibb := 0
		jbb := 0

		if firstShipI != nil {
			ibb = firstShipI.Battles
		}

		if firstShipJ != nil {
			jbb = firstShipJ.Battles
		}

		ib := shipSlice[i].Battles - ibb
		jb := shipSlice[j].Battles - jbb

		return ib > jb
	})

	getMostPlayedShip := func(ship *storage.StoredShip) (string, int) {
		firstShip := file.First.Ships[ship.ShipID]
		b := 0

		if firstShip != nil {
			b = firstShip.Battles
		}

		return wows.Ships[ship.ShipID].Name, ship.Battles - b
	}

	if len(shipSlice) > 0 {
		s.MostPlayedShip1, s.MostPlayedShip1Battles = getMostPlayedShip(shipSlice[0])
	}
	if len(shipSlice) > 1 {
		s.MostPlayedShip2, s.MostPlayedShip2Battles = getMostPlayedShip(shipSlice[1])
	}
	if len(shipSlice) > 2 {
		s.MostPlayedShip3, s.MostPlayedShip3Battles = getMostPlayedShip(shipSlice[2])
	}

	err = gocsv.MarshalWithoutHeaders([]*Statistics{&s}, os.Stdout)
	if err != nil {
		panic(err)
	}
}

func readFile(path string) (*MergedFile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	mf := new(MergedFile)
	if err := json.Unmarshal(data, mf); err != nil {
		return nil, err
	}

	return mf, nil
}
