package model

type FoodPetFoodPlan struct {
	ID            uint `gorm:"primaryKey;autoIncrement" json:"id"`
	PetFoodPlanID uint `gorm:"not null" json:"pet_food_plan_id"`
	FoodID        uint `gorm:"not null" json:"food_id"`

	// Relationships
	// PetFoodPlan       *PetFoodPlan       `gorm:"foreignKey:PetFoodPlanID" json:"pet_food_plan,omitempty"`
	Food              *Food              `gorm:"foreignKey:FoodID" json:"food,omitempty"`
	PetFoodPlanDetail *PetFoodPlanDetail `gorm:"foreignKey:FoodPetFoodPlanID" json:"pet_food_plan_detail,omitempty"`
}
