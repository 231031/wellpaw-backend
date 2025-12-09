package model

type DiseaseType int

const (
	RINGWORM DiseaseType = iota
	SCABIES
	HEALTHY
	DEMODICOSIS
	PYODERMA
)

var DiseaseTypeLabel = map[DiseaseType]string{
	RINGWORM:    "Ringworm",
	SCABIES:     "Scabies",
	DEMODICOSIS: "Demodicosis",
	PYODERMA:    "Pyoderma",
	HEALTHY:     "Healthy",
}

func (disease DiseaseType) String() string {
	return DiseaseTypeLabel[disease]
}
