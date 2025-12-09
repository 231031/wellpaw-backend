package model

import "time"

type PetFoodPlan struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PetID     uint      `json:"pet_id"`
	Name      string    `gorm:"type:varchar(255)" json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	// Pet                *Pet                 `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	FoodPetFoodPlans   []FoodPetFoodPlan    `gorm:"foreignKey:PetFoodPlanID" json:"food_pet_food_plans,omitempty"`
	PetFoodPlanHistory []PetFoodPlanHistory `gorm:"foreignKey:PetFoodPlanID" json:"pet_food_plan_history,omitempty"`
}
