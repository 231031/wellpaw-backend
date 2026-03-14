package model

type FoodPetFoodPlan struct {
	ID            uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	PetFoodPlanID uint    `gorm:"not null" json:"pet_food_plan_id"`
	FoodID        uint    `gorm:"not null" json:"food_id"`
	GramsPerCup   float64 `json:"grams_per_cup"`
	Cup           float64 `gorm:"-" json:"cups"`

	// Relationships
	PetFoodPlan        *PetFoodPlan        `gorm:"foreignKey:PetFoodPlanID" json:"pet_food_plan,omitempty"`
	Food               *Food               `gorm:"foreignKey:FoodID" json:"food,omitempty"`
	PetFoodPlanDetails []PetFoodPlanDetail `gorm:"foreignKey:FoodPetFoodPlanID" json:"pet_food_plan_detail,omitempty"`
}
