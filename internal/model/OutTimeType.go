package model

type OutTimeType int

const (
	NOW OutTimeType = iota
	NEXT
)

var OutTimeTypeLabel = map[OutTimeType]string{
	NOW:  "Now",
	NEXT: "Next Meal",
}

func (out OutTimeType) String() string {
	return OutTimeTypeLabel[out]
}
