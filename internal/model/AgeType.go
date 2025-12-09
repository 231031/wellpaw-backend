package model

type AgeType int

const (
	JUNIOR AgeType = iota
	ADULT
	SENIOR
)

// cat junior 2-12 months, senior 7+ years
// dogs
var AgeTypeLabel = map[AgeType]string{
	JUNIOR: "Junior",
	ADULT:  "Adult",
	SENIOR: "Senior",
}

func (agerange AgeType) String() string {
	return AgeTypeLabel[agerange]
}
