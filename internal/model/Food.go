package model

import "gorm.io/gorm"

type Food struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint           `gorm:"not null" json:"user_id" validate:"required"`
	ImagePath   string         `gorm:"type:varchar(512)" json:"image_path"`
	Name        string         `gorm:"not null;type:varchar(255)" json:"name" validate:"required"`
	Brand       string         `gorm:"not null;type:varchar(255)" json:"brand" validate:"required"`
	Type        *FoodType      `gorm:"not null" json:"type" validate:"required,oneof=0 1 2 3"`
	Energy      float64        `gorm:"not null" json:"energy" validate:"gt=-1"`
	Protein     float64        `gorm:"not null" json:"protein" validate:"gt=-1"`
	Fat         float64        `gorm:"not null" json:"fat" validate:"gt=-1"`
	Moist       float64        `gorm:"not null" json:"moist" validate:"gt=-1"`
	GramsPerCup float64        `json:"grams_per_cup"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Quantity    int     `gorm:"-" json:"quantity,omitempty" validate:"required,gt=0"`
	Weight      float64 `gorm:"-" json:"weight,omitempty" validate:"required,gt=0"`
	TotalAmount float64 `gorm:"-" json:"total_amount"`

	// Relationships
	User            *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	FoodPetFoodPlan []FoodPetFoodPlan `gorm:"foreignKey:FoodID" json:"food_pet_food_plans,omitempty"`
	FoodQuantities  []FoodQuantity    `gorm:"foreignKey:FoodID" json:"food_quantities,omitempty"`
}
