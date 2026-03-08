package model

type UsageFeatureType int

const (
	PROFILE UsageFeatureType = iota
	FOOD
	FOODPLAN
	DISEASE
)

var UsageFeatureLabel = map[UsageFeatureType]string{
	PROFILE:  "Profile",
	FOOD:     "Food",
	FOODPLAN: "Food Plan",
	DISEASE:  "Disease",
}

func (feature UsageFeatureType) String() string {
	return UsageFeatureLabel[feature]
}
