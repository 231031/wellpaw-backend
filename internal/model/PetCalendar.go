package model

import (
	"time"

	"gorm.io/gorm"
)

type PetCalendar struct {
	ID            uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string          `gorm:"not null;type:varchar(255)" json:"name" validate:"required"`
	Type          *CalendarType   `gorm:"not null" json:"type" validate:"required"`
	StartDatetime time.Time       `gorm:"not null" json:"start_datetime" validate:"required"`
	EndDate       time.Time       `gorm:"type:date" json:"end_date"`
	Frequently    *FrequentlyType `gorm:"not null" json:"frequently" validate:"required"`
	Notation      string          `gorm:"type:varchar(512)" json:"notation"`
	CreatedAt     time.Time       `gorm:"not null" json:"created_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	ActivityEvents []PetActivityCalendar `gorm:"foreignKey:PetCalendarID" json:"activity_events,omitempty"`
}
