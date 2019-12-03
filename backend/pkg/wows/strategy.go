package wows

// EventStrategy describes an interface to cover any whaling event, as they all contain different logic
type EventStrategy interface {
	IsShipEligible(*Warship) bool
	GetShipRedeemable(*Warship) (*Resource, *uint)
}
