package model

import "time"

type PetSkinImage struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PetID         uint      `json:"pet_id"`
	ImagePath     string    `gorm:"type:varchar(255)" json:"image_path"`
	Predicted     int       `json:"predicted"`
	Confident     int       `json:"confident"`
	Labeled       int       `json:"labeled"`
	ImageEvidence string    `gorm:"type:varchar(512)" json:"image_evidence"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	// Pet *Pet `gorm:"foreignKey:PetID" json:"pet,omitempty"`
}
