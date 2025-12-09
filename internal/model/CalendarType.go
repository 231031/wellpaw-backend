package model

type CalendarType int

const (
	VACCINE CalendarType = iota
	DRUG
	APPOINTMENT
)

var CalendarTypeLabel = map[CalendarType]string{
	VACCINE:     "Vaccine",
	DRUG:        "Drug",
	APPOINTMENT: "Appointment",
}

func (food CalendarType) String() string {
	return CalendarTypeLabel[food]
}
