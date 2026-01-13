package service

import (
	"context"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/stripe/stripe-go/v84"
)

type PaymentService interface {
	CreateCustomer(ctx context.Context, user *model.User) (string, error)
	AttachPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) error
}

type paymentService struct {
	stripeClient *stripe.Client
}

func NewPaymentService(stripeClient *stripe.Client) PaymentService {
	return &paymentService{
		stripeClient: stripeClient,
	}
}

func (s *paymentService) CreateCustomer(ctx context.Context, user *model.User) (string, error) {
	params := &stripe.CustomerCreateParams{
		Name:  stripe.String(user.FirstName + " " + user.LastName),
		Email: stripe.String(user.Email),
	}

	c, err := s.stripeClient.V1Customers.Create(ctx, params)
	if err != nil {
		return "", err
	}
	return c.ID, nil
}

func (s *paymentService) AttachPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) error {
	// frontend : create payment method
	_, err := s.stripeClient.V1PaymentMethods.Attach(ctx, paymentMethodID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	})
	if err != nil {
		return err
	}

	return nil
}
