package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/stripe/stripe-go/v84"
)

type WebhookService interface {
	HandleSubscriptionUpdated(ctx context.Context, eventObject *stripe.Event) error
}

type webhookService struct {
	userRepository repository.UserRepository
}

func NewWebhookService(userRepository repository.UserRepository) WebhookService {
	return &webhookService{
		userRepository: userRepository,
	}
}

func (s *webhookService) HandleSubscriptionUpdated(ctx context.Context, eventObject *stripe.Event) error {
	subscription, err := s.mapSubscriptionDetail(eventObject.Data.Object)
	if err != nil {
		return err
	}

	status := utils.ConvertSubscriptionStatusToSubscriptionStatusType(*stripe.String(subscription.Status))
	tier := utils.ConvertIntervalToTierType(*stripe.String(subscription.Items.Data[0].Plan.Interval))

	if err := s.userRepository.UpdateSubscriptionDetail(ctx, subscription.Customer.ID, status, tier); err != nil {
		return err
	}
	return nil
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
