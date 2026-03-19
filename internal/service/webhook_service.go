package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/stripe/stripe-go/v84"
)

type WebhookService interface {
	MapWebhookHandler(ctx context.Context, eventObject *stripe.Event) error
	handleSubscriptionUpdated(ctx context.Context, eventObject *stripe.Event) error
	handleSubscriptionFreeTierUpdated(ctx context.Context, eventObject *stripe.Event) error
}

type webhookService struct {
	fcmService     FcmService
	userRepository repository.UserRepository
}

func NewWebhookService(userRepository repository.UserRepository, fcmService FcmService) WebhookService {
	return &webhookService{
		fcmService:     fcmService,
		userRepository: userRepository,
	}
}

func (s *webhookService) MapWebhookHandler(ctx context.Context, eventObject *stripe.Event) error {
	switch eventObject.Type {
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, eventObject)
	case "customer.subscription.deleted":
		return s.handleSubscriptionUpdated(ctx, eventObject)
	case "invoice.payment_succeeded":
		return s.handleSubscriptionFreeTierUpdated(ctx, eventObject)
	default:
		applogger.LogInfo(fmt.Sprintf("webhook doesn't handle this event type %s", *stripe.String(eventObject.Type)), serviceLog)
	}

	return nil
}

func (s *webhookService) handleSubscriptionUpdated(ctx context.Context, eventObject *stripe.Event) error {
	subscription, err := s.mapSubscriptionDetail(eventObject.Data.Object)
	if err != nil {
		return err
	}

	status := utils.ConvertSubscriptionStatusToSubscriptionStatusType(*stripe.String(subscription.Status))
	tier := utils.ConvertIntervalToTierType(*stripe.String(subscription.Items.Data[0].Plan.Interval))

	user, err := s.userRepository.UpdateSubscriptionDetail(ctx, subscription.Customer.ID, status, tier)
	if err != nil {
		applogger.LogError(fmt.Sprintf("webhook handler failed to update subscription detail : %s", err.Error()), serviceLog)
		return err
	}

	err = s.userRepository.SetCurrentSubscriptionDetail(ctx, user.ID, tier, status)
	if err != nil {
		applogger.LogError(fmt.Sprintf("webhook handler failed to set subscription detail in redsi : %s", err.Error()), serviceLog)
		return err
	}

	s.fcmService.SendSilentSubscriptionNotification(ctx, user.DeviceToken, tier, status)

	return nil
}

func (s *webhookService) handleSubscriptionFreeTierUpdated(ctx context.Context, eventObject *stripe.Event) error {
	invoice, err := s.mapFreeTrialInvoice(eventObject.Data.Object)
	if err != nil {
		return err
	}

	if len(invoice.Lines.Data) > 0 {
		if invoice.Lines.Data[0].Amount != 0 {
			return nil
		}
	} else {
		applogger.LogInfo(fmt.Sprintf("%s doesn't have data in lines", invoice.ID), serviceLog)
		return nil
	}

	user, err := s.userRepository.UpdateSubscriptionDetail(ctx, invoice.Customer.ID, model.ACTIVESUB, model.FREE)
	if err != nil {
		applogger.LogError(fmt.Sprintf("webhook handler failed to update subscription free-tial detail : %s", err.Error()), serviceLog)
		return err
	}

	err = s.userRepository.SetCurrentSubscriptionDetail(ctx, user.ID, model.FREE, model.ACTIVESUB)
	if err != nil {
		applogger.LogError(fmt.Sprintf("webhook handler failed to set subscription free-tial detail in redsi : %s", err.Error()), serviceLog)
		return err
	}

	s.fcmService.SendSilentSubscriptionNotification(ctx, user.DeviceToken, model.FREE, model.ACTIVESUB)

	return nil
}

func (s *webhookService) mapFreeTrialInvoice(object map[string]interface{}) (stripe.Invoice, error) {
	var invoice stripe.Invoice
	dataBytes, err := json.Marshal(object)
	if err != nil {
		return invoice, fmt.Errorf("error marshaling event data: %v", err)
	}

	err = json.Unmarshal(dataBytes, &invoice)
	if err != nil {
		return invoice, fmt.Errorf("error unmarshaling to invoice: %v", err)
	}

	return invoice, nil
}

func (s *webhookService) mapSubscriptionDetail(object map[string]interface{}) (stripe.Subscription, error) {
	var subscription stripe.Subscription
	dataBytes, err := json.Marshal(object)
	if err != nil {
		return subscription, fmt.Errorf("error marshaling event data: %v", err)
	}

	err = json.Unmarshal(dataBytes, &subscription)
	if err != nil {
		return subscription, fmt.Errorf("error unmarshaling to subscription: %v", err)
	}

	return subscription, nil
}
