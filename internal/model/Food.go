package model

type Food struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id" validate:"required"`
	ImagePath string    `gorm:"type:varchar(512)" json:"image_path"`
	Name      string    `gorm:"not null;type:varchar(255)" json:"name" validate:"required"`
	Brand     string    `gorm:"not null;type:varchar(255)" json:"brand" validate:"required"`
	Type      *FoodType `gorm:"not null" json:"type" validate:"required,oneof=0 1 2 3"`
	Quantity  int       `gorm:"not null" json:"quantity" validate:"required,gt=0"`
	Weight    float64   `gorm:"not null" json:"weight" validate:"required,gt=0"`
	Energy    float64   `gorm:"not null" json:"energy" validate:"required,gt=-1"`
	Protein   float64   `gorm:"not null" json:"protein" validate:"required,gt=-1"`
	Fat       float64   `gorm:"not null" json:"fat" validate:"required,gt=-1"`
	Moist     float64   `gorm:"not null" json:"moist" validate:"required,gt=-1"`

	// Relationships
	User            *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	FoodPetFoodPlan []FoodPetFoodPlan `gorm:"foreignKey:FoodID" json:"food_pet_food_plans,omitempty"`
}
