package model

import (
	"time"

	"gorm.io/gorm"
)

type PetFoodPlan struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	PetID     uint           `gorm:"not null" json:"pet_id"`
	Name      string         `gorm:"not null;type:varchar(255)" json:"name"`
	Active    bool           `gorm:"not null" json:"active"`
	Unit      UnitType       `gorm:"default:0;not null" json:"unit"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	// Pet                *Pet                 `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	FoodPetFoodPlans  []FoodPetFoodPlan  `gorm:"foreignKey:PetFoodPlanID" json:"food_pet_food_plans,omitempty"`
	PetFoodPlanTotals []PetFoodPlanTotal `gorm:"foreignKey:PetFoodPlanID" json:"pet_food_plan_totals,omitempty"`
}
