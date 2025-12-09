package model

type SexType int

const (
	FEMALE SexType = iota
	MALE
)

var SexTypeLabel = map[SexType]string{
	FEMALE: "Female",
	MALE:   "Male",
}

func (sex SexType) String() string {
	return SexTypeLabel[sex]
}
