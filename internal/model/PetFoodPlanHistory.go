package model

import "time"

type PetFoodPlanHistory struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PetID              uint      `gorm:"not null" json:"pet_id"`
	PetFoodPlanTotalID uint      `gorm:"not null" json:"pet_food_plan_total_id"`
	PetDetailID        uint      `gorm:"not null" json:"pet_detail_id"`
	CreatedAt          time.Time `gorm:"not null" json:"created_at"`
	PlanUsageEndDate   time.Time `gorm:"-" json:"plan_usage_end_date,omitempty"`

	// Relationships
	// Pet              *Pet              `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	PetFoodPlanTotal *PetFoodPlanTotal `gorm:"foreignKey:PetFoodPlanTotalID;references:ID" json:"pet_food_plan_total,omitempty"`
	PetDetail        *PetDetail        `gorm:"foreignKey:PetDetailID" json:"pet_detail,omitempty"`
}

func (PetFoodPlanHistory) TableName() string {
	return "pet_food_plan_histories"
}
