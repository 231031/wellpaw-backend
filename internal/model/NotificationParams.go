package model

type SendNotificationParams struct {
	Token string
	Title string
	Body  string
	Data  map[string]string
}
