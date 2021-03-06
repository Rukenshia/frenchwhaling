package wows

//go:generate go run ./gen/generate_warships.go

import (
	"strings"
)

// Warship represents a WoWS Ship with field relevant to Whaling
type Warship struct {
	Name           string           `json:"name"`
	PriceGold      int              `json:"price_gold"`
	Nation         string           `json:"nation"`
	IsPremium      bool             `json:"is_premium"`
	ShipID         int64            `json:"ship_id"`
	PriceCredit    int              `json:"price_credit"`
	Tier           int              `json:"tier"`
	NextShips      map[string]int64 `json:"next_ships"`
	HasDemoProfile bool             `json:"has_demo_profile"`
}

// IsRentalShip returns whether a ship is only available for a limited period of time,
// such as ships for rent events or clan battles
func (w *Warship) IsRentalShip() bool {
	return strings.Contains(w.Name, "[")
}

// IsTestShip returns whether the ship is currently in testing (WIP ships)
func (w *Warship) IsTestShip() bool {
	return w.HasDemoProfile
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
