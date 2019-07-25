package wows

//go:generate go run ./gen/generate_warships.go

import (
	"strings"
)

// Warship represents a WoWS Ship with field relevant to Frenchwhaling
type Warship struct {
	Name        string           `json:"name"`
	PriceGold   int              `json:"price_gold"`
	Nation      string           `json:"nation"`
	IsPremium   bool             `json:"is_premium"`
	ShipID      int64            `json:"ship_id"`
	PriceCredit int              `json:"price_credit"`
	Tier        int              `json:"tier"`
	NextShips   map[string]int64 `json:"next_ships"`
}

// IsEgligible returns whether the ship can participate in the event
func (w *Warship) IsEgligible() bool {
	if strings.Contains(w.Name, "[") {
		return false
	}

	if w.IsTestShip() {
		return false
	}

	if w.Tier < 5 && !w.GetsPremiumTreatment() {
		return false
	}

	if w.Tier < 2 {
		return false
	}

	return true
}

// IsTestShip returns whether the ship is currently in testing (WIP ships)
func (w *Warship) IsTestShip() bool {
	for _, ts := range []string{
		"Hayate", "Friesland", "Ark Royal", "Siegfried",
		"Ohio", "Somers", "Colbert", "Smolensk", "Slava",
		"Thunderer", "Kléber",
	} {
		if w.Name == ts {
			return true
		}
	}

	return false
}

// GetsPremiumTreatment returns whether the ship is a premium or premium in disguise, like Armory ship,
// which are treated as Premium ships in the event
func (w *Warship) GetsPremiumTreatment() bool {
	if w.IsPremium {
		return true
	}

	// ARP Event ships
	if strings.Contains(w.Name, "ARP ") {
		return true
	}

	// Armory premiums (non-T10)
	if len(w.NextShips) == 0 && w.Tier < 10 {
		return true
	}

	if w.PriceCredit == 0 {
		return true
	}

	return false
}

// Resource returns the Resource the ship will yield
func (w *Warship) Resource() Resource {
	if w.Tier < 7 {
		return Coal
	}
	return RepublicTokens
}

// Amount returns the quantity of the resource the ship will yield
func (w *Warship) Amount() uint {
	switch w.GetsPremiumTreatment() {
	case true:
		switch w.Tier {
		case 10:
			fallthrough
		case 9:
			return 20
		case 8:
			return 15
		case 7:
			return 10
		case 6:
			return 400
		case 5:
			return 300
		default:
			return 200
		}
	case false:
		switch w.Tier {
		case 10:
			fallthrough
		case 9:
			return 15
		case 8:
			return 10
		case 7:
			return 5
		case 6:
			return 300
		default:
			return 200
		}
	}

	return 0
}
