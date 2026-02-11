package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	mailjet "github.com/mailjet/mailjet-apiv3-go/v4"
)

type EmailService interface {
	SendOTPEmail(ctx context.Context, receiverEmail string, otp string) error
}

type emailService struct {
	client      *mailjet.Client
	senderEmail string
	senderName  string
}

func NewEmailService(apiKey, apiSecret, senderEmail, senderName string) EmailService {
	client := mailjet.NewMailjetClient(apiKey, apiSecret)
	client.SetClient(&http.Client{
		Timeout: 10 * time.Second,
	})

	return &emailService{
		client:      client,
		senderEmail: senderEmail,
		senderName:  senderName,
	}
}

func (s *emailService) SendOTPEmail(ctx context.Context, receiverEmail string, otp string) error {
	toRecipients := mailjet.RecipientsV31{
		{
			Email: receiverEmail,
		},
	}

	message := mailjet.MessagesV31{
		Info: []mailjet.InfoMessagesV31{
			{
				From: &mailjet.RecipientV31{
					Email: s.senderEmail,
					Name:  s.senderName,
				},
				To:       &toRecipients,
				Subject:  "WellPaw OTP Verification",
				TextPart: fmt.Sprintf("Your verification code is %s. The code will expire in 5 minutes.", otp),
			},
		},
	}

	result, err := s.client.SendMailV31(&message, mailjet.WithContext(ctx))
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to call mailjet sdk: %s", err.Error()), serviceLog)
		return fmt.Errorf("failed to call mailjet sdk: %w", err)
	}

	if result == nil || len(result.ResultsV31) == 0 {
		return fmt.Errorf("mailjet sdk returned empty response")
	}

	if !strings.EqualFold(result.ResultsV31[0].Status, "success") {
		applogger.LogError(fmt.Sprintf("mailjet send email status is %s", result.ResultsV31[0].Status), serviceLog)
		return fmt.Errorf("mailjet send email failed with status %s", result.ResultsV31[0].Status)
	}

	return nil
}
