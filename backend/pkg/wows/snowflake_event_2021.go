package wows

type Snowflake2021 struct {
}

func (s Snowflake2021) IsShipEligible(w *Warship) bool {
	if w.IsTestShip() {
		return false
	}

	if w.IsRentalShip() {
		return false
	}

	if w.Tier < 5 {
		return false
	}

	return true
}

func (s Snowflake2021) GetShipRedeemable(w *Warship) (Resource, uint) {
	switch w.Tier {
	case 5:
		fallthrough
	case 6:
		fallthrough
	case 7:
		return Coal, 750
	case 8:
		fallthrough
	case 9:
		return Steel, 75
	case 10:
		return NewYearCertificate, 1
	default:
		return 0, 0
	}
}
