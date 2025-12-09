package model

type BcsType int

const (
	VERYTHIN BcsType = iota
	THIN
	IDEAL
	OVERWEIGHT
	OBESITY
)

var BcsTypeLabel = map[BcsType]string{
	VERYTHIN:   "Very Thin",
	THIN:       "Thin",
	IDEAL:      "Ideal",
	OVERWEIGHT: "Overweight",
	OBESITY:    "Obesity",
}

func (pet BcsType) String() string {
	return BcsTypeLabel[pet]
}
