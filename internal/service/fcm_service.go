package service

import (
	"context"

	"firebase.google.com/go/v4/messaging"
	"github.com/231031/wellpaw-backend/internal/model"
)

type FcmService interface {
	SendNotifications(ctx context.Context, params []model.SendNotificationParams) (*messaging.BatchResponse, error)
}

type fcmService struct {
	fcmClient *messaging.Client
}

func NewFCMService(fcmClient *messaging.Client) *fcmService {
	return &fcmService{fcmClient: fcmClient}
}

func (s *fcmService) SendNotifications(
	ctx context.Context,
	params []model.SendNotificationParams,
) (*messaging.BatchResponse, error) {

	messages := make([]*messaging.Message, 0, len(params))

	for _, p := range params {

		msg := &messaging.Message{
			Token: p.Token,

			Notification: &messaging.Notification{
				Title: p.Title,
				Body:  p.Body,
			},

			Data: p.Data,

			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID: "reminder_channel",
					Sound:     "default",
				},
			},

			APNS: &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-priority": "10",
				},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Sound: "default",
					},
				},
			},
		}

		messages = append(messages, msg)
	}

	resp, err := s.fcmClient.SendEach(ctx, messages)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
