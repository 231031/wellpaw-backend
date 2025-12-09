package model

import "time"

type PetFoodPlanHistory struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PetFoodPlanID uint      `gorm:"not null" json:"pet_food_plan_id"`
	PetDetailID   uint      `gorm:"not null" json:"pet_detail_id"`
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`

	// Relationships
	// PetFoodPlan *PetFoodPlan `gorm:"foreignKey:PetFoodPlanID" json:"pet_food_plan,omitempty"`
	// PetDetail   *PetDetail   `gorm:"foreignKey:PetDetailID" json:"pet_detail,omitempty"`
}
