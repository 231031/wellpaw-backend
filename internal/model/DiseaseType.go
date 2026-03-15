package model

type DiseaseType int

const (
	RINGWORM DiseaseType = iota
	SCABIES
	HEALTHY
	DEMODICOSIS
	PYODERMA
	OTHER
	BACTERIAL_DERMATOSIS
)

var DiseaseTypeLabel = map[DiseaseType]string{
	RINGWORM:             "Ringworm",
	SCABIES:              "Scabies",
	DEMODICOSIS:          "Demodicosis",
	PYODERMA:             "Pyoderma",
	HEALTHY:              "Healthy",
	OTHER:                "Other",
	BACTERIAL_DERMATOSIS: "Bacterial Dermatosis",
}

func (disease DiseaseType) String() string {
	return DiseaseTypeLabel[disease]
}
