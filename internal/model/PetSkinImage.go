package model

import "time"

type PetSkinImage struct {
	ID            uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	PetID         uint        `gorm:"not null" json:"pet_id"`
	ImagePath     string      `gorm:"not null;type:varchar(255)" json:"image_path"`
	Predicted     DiseaseType `gorm:"not null" json:"predicted"`
	Confident     int         `gorm:"not null" json:"confident"`
	Labeled       DiseaseType `json:"labeled"`
	ImageEvidence string      `gorm:"type:varchar(512)" json:"image_evidence"`
	CreatedAt     time.Time   `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time   `gorm:"not null" json:"updated_at"`

	// Relationships
	// Pet *Pet `gorm:"foreignKey:PetID" json:"pet,omitempty"`
}
