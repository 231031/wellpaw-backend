package service

import (
	"context"
	"fmt"
	"strconv"

	"firebase.google.com/go/v4/messaging"
	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
)

type FcmService interface {
	SendNotifications(ctx context.Context, params []model.SendNotificationParams) (*messaging.BatchResponse, error)
	SendSilentSubscriptionNotification(ctx context.Context, token string, tier model.TierType, status model.SubscriptionStatusType) (string, error)
}

type fcmService struct {
	fcmClient *messaging.Client
}

const maxFCMSendEachBatch = 300

func NewFCMService(fcmClient *messaging.Client) FcmService {
	return &fcmService{fcmClient: fcmClient}
}

func (s *fcmService) SendNotifications(
	ctx context.Context,
	params []model.SendNotificationParams,
) (*messaging.BatchResponse, error) {
	if len(params) == 0 {
		return &messaging.BatchResponse{
			SuccessCount: 0,
			FailureCount: 0,
			Responses:    []*messaging.SendResponse{},
		}, nil
	}

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

	if len(messages) <= maxFCMSendEachBatch {
		resp, err := s.fcmClient.SendEach(ctx, messages)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	combined := &messaging.BatchResponse{
		SuccessCount: 0,
		FailureCount: 0,
		Responses:    make([]*messaging.SendResponse, 0, len(messages)),
	}

	for start := 0; start < len(messages); start += maxFCMSendEachBatch {
		end := start + maxFCMSendEachBatch
		if end > len(messages) {
			end = len(messages)
		}

		resp, err := s.fcmClient.SendEach(ctx, messages[start:end])
		if err != nil {
			return nil, err
		}

		combined.SuccessCount += resp.SuccessCount
		combined.FailureCount += resp.FailureCount
		combined.Responses = append(combined.Responses, resp.Responses...)
	}

	return combined, nil
}

func (s *fcmService) SendSilentSubscriptionNotification(
	ctx context.Context,
	token string,
	tier model.TierType,
	status model.SubscriptionStatusType,
) (string, error) {

	msg := &messaging.Message{
		Token: token,

		Data: map[string]string{
			"type":                      "subscription_update",
			"tier":                      strconv.Itoa(int(tier)),
			"subscription_status":       strconv.Itoa(int(status)),
			"tier_label":                tier.String(),
			"subscription_status_label": status.String(),
		},

		Android: &messaging.AndroidConfig{
			Priority: "normal",
		},

		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority":  "5",
				"apns-push-type": "background",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: true,
				},
			},
		},
	}

	messageID, err := s.fcmClient.Send(ctx, msg)
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to send silent notification : %v", err), serviceLog)
		return "", err
	}

	return messageID, nil
}
