package model

type ActivityLevel int

const (
	INACTIVE ActivityLevel = iota
	SOMEACTIVE
	ACTIVE
	VERYACTIVE
)

var ActivityLevelLabel = map[ActivityLevel]string{
	INACTIVE:   "Inactive",
	SOMEACTIVE: "Somewhat Active",
	ACTIVE:     "Active",
	VERYACTIVE: "Very active",
}

func (activity ActivityLevel) String() string {
	return ActivityLevelLabel[activity]
}
