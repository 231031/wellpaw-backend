package model

type PetType int

const (
	DOG PetType = iota
	CAT
)

var PetTypeLabel = map[PetType]string{
	DOG: "Dog",
	CAT: "Cat",
}

func (pet PetType) String() string {
	return PetTypeLabel[pet]
}
