package model

type FrequentlyType int

const (
	NOT FrequentlyType = iota
	DAY
	WEEK
	MONTH
	YEAR
)

var FrequentlyTypeLabel = map[FrequentlyType]string{
	NOT:   "Not",
	DAY:   "Day",
	WEEK:  "Week",
	MONTH: "Month",
	YEAR:  "Year",
}

func (frequently FrequentlyType) String() string {
	return FrequentlyTypeLabel[frequently]
}
