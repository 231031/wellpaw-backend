package model

type ReproductionType int

const (
	INTACT ReproductionType = iota
	NEUTERED
)

var ReproductionTypeLabel = map[ReproductionType]string{
	INTACT:   "Intact",
	NEUTERED: "Neutered",
}

func (reproduction ReproductionType) String() string {
	return ReproductionTypeLabel[reproduction]
}
