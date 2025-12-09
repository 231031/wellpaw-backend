package model

import "time"

type PetCalendar struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string         `gorm:"type:varchar(255)" json:"name"`
	Type          int            `json:"type"`
	StartDatetime time.Time      `json:"start_datetime"`
	EndDate       time.Time      `gorm:"type:date" json:"end_date"`
	Frequently    FrequentlyType `json:"frequently"`
	Notation      string         `gorm:"type:varchar(512)" json:"notation"`
	CreatedAt     time.Time      `json:"created_at"`

	// Relationships
	ActivityEvents []PetActivityCalendar `gorm:"foreignKey:PetCalendarID" json:"activity_events,omitempty"`
}
