package wows

type BirthdayEvent2021 struct {
}

func (e BirthdayEvent2021) IsShipEligible(w *Warship) bool {
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

func (e BirthdayEvent2021) GetShipRedeemable(w *Warship) (Resource, uint) {
	switch w.Tier {
	case 5:
		fallthrough
	case 6:
		fallthrough
	case 7:
		return FestiveToken, 1
	case 8:
		fallthrough
	case 9:
		return FestiveTokenAndAnniversaryContainer, 1
	case 10:
		return SuperContainer, 1
	default:
		return 0, 0
	}
}
