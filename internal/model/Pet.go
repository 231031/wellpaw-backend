package model

import "time"

type Pet struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id" validate:"required"`
	ImagePath string    `gorm:"type:varchar(512)" json:"image_path"`
	Name      string    `gorm:"not null;type:varchar(255)" json:"name" validate:"required"`
	Type      *PetType  `gorm:"not null" json:"type" validate:"required,oneof=0 1"`
	Breed     string    `gorm:"not null;type:varchar(255)" json:"breed" validate:"required"`
	SexType   *SexType  `gorm:"not null" json:"sex_type" validate:"required,oneof=0 1"`
	BirthDate time.Time `gorm:"not null;type:date" json:"birth_date" validate:"required"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`

	// Relationships
	// User           *User                 `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PetDetails          []PetDetail           `gorm:"foreignKey:PetID" json:"pet_details,omitempty"`
	PetFoodPlans        []PetFoodPlan         `gorm:"foreignKey:PetID" json:"pet_food_plans,omitempty"`
	PetFoodPlanHistorys []PetFoodPlanHistory  `gorm:"foreignKey:PetID" json:"pet_food_plan_historys,omitempty"`
	PetSkinImages       []PetSkinImage        `gorm:"foreignKey:PetID" json:"pet_skin_images,omitempty"`
	ActivityEvents      []PetActivityCalendar `gorm:"foreignKey:PetID" json:"activity_events,omitempty"`
}
