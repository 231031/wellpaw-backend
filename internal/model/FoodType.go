package model

type FoodType int

const (
	DRY FoodType = iota
	WET
	TREATS
	SUPPLEMENTS
)

var FoodTypeLabel = map[FoodType]string{
	DRY:         "Dry",
	WET:         "Wet",
	TREATS:      "Treats",
	SUPPLEMENTS: "Supplements",
}

func (food FoodType) String() string {
	return FoodTypeLabel[food]
}
