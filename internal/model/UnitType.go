package model

type UnitType int

const (
	GRAM UnitType = iota
	CUP
)

var UnitTypeLabel = map[UnitType]string{
	GRAM: "Gram",
	CUP:  "Cup",
}

func (unit UnitType) String() string {
	return UnitTypeLabel[unit]
}
