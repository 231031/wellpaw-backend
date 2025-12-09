package model

type TierType int

const (
	FREE TierType = iota
	WEEKS
	MONTHS
	YEARS
)

var TierTypeLabel = map[TierType]string{
	FREE:   "Dry",
	WEEKS:  "Week",
	MONTHS: "Month",
	YEARS:  "Year",
}

func (tier TierType) String() string {
	return TierTypeLabel[tier]
}
