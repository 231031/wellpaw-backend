package model

type PetActivityCalendar struct {
	ID            uint `gorm:"primaryKey;autoIncrement" json:"id"`
	PetID         uint `gorm:"not null" json:"pet_id"`
	PetCalendarID uint `gorm:"not null" json:"pet_calendar_id"`

	// Relationships
	Pet         *Pet         `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	PetCalendar *PetCalendar `gorm:"foreignKey:PetCalendarID" json:"pet_calendar,omitempty"`
}
