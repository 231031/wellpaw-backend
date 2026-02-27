package model

import "time"

type FoodQuantity struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FoodID    uint      `gorm:"not null" json:"food_id"`
	Quantity  int       `gorm:"not null" json:"quantity" validate:"required,gt=0"`
	Weight    float64   `gorm:"not null" json:"weight" validate:"required,gt=0"`
	Amount    float64   `gorm:"not null" json:"amount"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`

	Food *Food `gorm:"foreignKey:FoodID" json:"food,omitempty"`
}

func (FoodQuantity) TableName() string {
	return "food_quantities"
}
