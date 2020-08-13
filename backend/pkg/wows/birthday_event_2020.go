package wows

type BirthdayEvent2020 struct {
}

func (e BirthdayEvent2020) IsShipEligible(w *Warship) bool {
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

func (e BirthdayEvent2020) GetShipRedeemable(w *Warship) (Resource, uint) {
	switch w.Tier {
	case 5:
		fallthrough
	case 6:
		fallthrough
	case 7:
		return AnniversaryCamouflages, 2
	case 8:
		return AnniversaryContainers, 1
	case 9:
		return AnniversaryContainers, 2
	case 10:
		return SuperContainer, 1
	default:
		return 0, 0
	}
}
