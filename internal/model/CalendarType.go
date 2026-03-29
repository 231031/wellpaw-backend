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

var CalendarTypeThaiLabel = map[CalendarType]string{
	VACCINE:     "ฉีดวัคซีน",
	DRUG:        "ให้ยา",
	APPOINTMENT: "นัดหมาย",
}

func (food CalendarType) StringThai() string {
	return CalendarTypeThaiLabel[food]
}
