package model

type NotificationPlan struct {
	Plan    PetFoodPlan
	OutTime OutTimeType
}

type SendNotificationParams struct {
	Token string
	Title string
	Body  string
	Data  map[string]string
}
