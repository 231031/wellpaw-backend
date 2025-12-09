package model

import "time"

type PetCalendar struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string         `gorm:"not null;type:varchar(255)" json:"name"`
	Type          CalendarType   `gorm:"not null" json:"type"`
	StartDatetime time.Time      `gorm:"not null" json:"start_datetime"`
	EndDate       time.Time      `gorm:"type:date" json:"end_date"`
	Frequently    FrequentlyType `gorm:"not null" json:"frequently"`
	Notation      string         `gorm:"type:varchar(512)" json:"notation"`
	CreatedAt     time.Time      `gorm:"not null" json:"created_at"`

	// Relationships
	ActivityEvents []PetActivityCalendar `gorm:"foreignKey:PetCalendarID" json:"activity_events,omitempty"`
}
