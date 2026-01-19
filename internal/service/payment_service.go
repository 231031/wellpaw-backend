package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/stripe/stripe-go/v84"
)

type PaymentService interface {
	CreateCustomer(ctx context.Context, user *model.User) (string, error)
	AttachPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) error

	GetAllSubscriptionsPlan(ctx context.Context) ([]*model.SubscriptionPlan, error)
	GetAllSubscriptionsByCustomerID(ctx context.Context, customerID string) ([]*model.SubscriptionHistory, error)
	GetPaymentIntentByID(ctx context.Context, paymentIntentID string) (*model.PaymentInvoice, error)
	GetSubscriptionScheduleByCustomerID(ctx context.Context, customerID string) ([]*stripe.SubscriptionSchedule, error)
	CreateSubscription(ctx context.Context, customerID, paymentMethodID, subscriptionPlanID string) *model.HTTPResponse
	UpdateSubscription(ctx context.Context, customerID, newSubscriptionPlanID string) *model.HTTPResponse
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

func (s *paymentService) GetAllSubscriptionsPlan(ctx context.Context) ([]*model.SubscriptionPlan, error) {
	params := &stripe.PriceListParams{
		Type:     stripe.String(string(stripe.PriceTypeRecurring)),
		Active:   stripe.Bool(true),
		Currency: stripe.String("thb"),
		Expand:   []*string{stripe.String("data.product")},
	}

	var subscriptionPlans []*model.SubscriptionPlan
	subscriptionPrices := s.stripeClient.V1Prices.List(ctx, params)

	for sp, err := range subscriptionPrices {
		if err != nil {
			return nil, err
		}

		if sp.Product != nil && !sp.Product.Active {
			continue
		}

		features := []string{}
		for _, feature := range sp.Product.MarketingFeatures {
			features = append(features, feature.Name)
		}

		subscriptionPlans = append(subscriptionPlans, &model.SubscriptionPlan{
			ID:            sp.ID,
			Name:          sp.Product.Name,
			Features:      features,
			Amount:        sp.UnitAmount / 100,
			Currency:      *stripe.String(sp.Currency),
			Interval:      *stripe.String(sp.Recurring.Interval),
			IntervalCount: int(*stripe.Int64(sp.Recurring.IntervalCount)),
		})
	}

	return subscriptionPlans, nil
}

func (s *paymentService) GetAllSubscriptionsByCustomerID(ctx context.Context, customerID string) ([]*model.SubscriptionHistory, error) {
	subscriptions := s.stripeClient.V1Subscriptions.List(ctx, &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
		Status:   stripe.String("all"),
		Expand: []*string{
			stripe.String("data.items.data"),
			stripe.String("data.items.data.price"),
			stripe.String("data.latest_invoice.lines.data"),
			stripe.String("data.latest_invoice.payments.data"),
		},
	})

	var subscriptionHistories []*model.SubscriptionHistory
	for sub, err := range subscriptions {
		if err != nil {
			return nil, err
		}

		var paymentIntentID string
		if len(sub.LatestInvoice.Payments.Data) > 0 {
			paymentIntentID = sub.LatestInvoice.Payments.Data[0].Payment.PaymentIntent.ID
		}

		subscriptionHistories = append(subscriptionHistories, &model.SubscriptionHistory{
			SubscriptionID:     sub.ID,
			SubscriptionStatus: string(sub.Status),
			InvoiceID:          sub.LatestInvoice.ID,
			InvoiceStatus:      string(sub.LatestInvoice.Status),
			PaymentIntentID:    paymentIntentID,
			PriceID:            sub.Items.Data[0].Price.ID,
			AmountPaid:         sub.LatestInvoice.AmountPaid / 100,
			AmountDue:          sub.LatestInvoice.AmountDue / 100,
			Amount:             sub.Items.Data[0].Price.UnitAmount / 100,
			PeriodStart:        utils.ConvertStripeTimeToTimeStr(sub.LatestInvoice.Lines.Data[0].Period.Start),
			PeriodEnd:          utils.ConvertStripeTimeToTimeStr(sub.LatestInvoice.Lines.Data[0].Period.End),
			Tier:               utils.ConvertIntervalToTierType(*stripe.String(sub.Items.Data[0].Price.Recurring.Interval)),
		})

	}

	return subscriptionHistories, nil
}

func (s *paymentService) GetPaymentIntentByID(ctx context.Context, paymentIntentID string) (*model.PaymentInvoice, error) {
	paymentIntent, err := s.stripeClient.V1PaymentIntents.Retrieve(ctx, paymentIntentID, &stripe.PaymentIntentRetrieveParams{
		Expand: []*string{
			stripe.String("customer.subscriptions.data"),
		},
	})
	if err != nil {
		return nil, err
	}

	paymentInvoice := &model.PaymentInvoice{
		ClientSecret:        paymentIntent.ClientSecret,
		PaymentIntentStatus: string(paymentIntent.Status),
		SubscriptionStatus:  string(paymentIntent.Customer.Subscriptions.Data[0].Status),
		Amount:              paymentIntent.Amount / 100,
	}
	return paymentInvoice, nil
}

func (s *paymentService) GetSubscriptionScheduleByCustomerID(ctx context.Context, customerID string) ([]*stripe.SubscriptionSchedule, error) {
	subSchedules := s.stripeClient.V1SubscriptionSchedules.List(ctx, &stripe.SubscriptionScheduleListParams{
		Customer: stripe.String(customerID),
		Expand: []*string{
			stripe.String("data.phases.items.price"),
			stripe.String("data.subscription.items.data"),
			stripe.String("data.subscription.latest_invoice.payments"),
		},
	})

	var allSubSchedules []*stripe.SubscriptionSchedule
	for subSchedule, err := range subSchedules {
		if err != nil {
			return nil, err
		}
		allSubSchedules = append(allSubSchedules, subSchedule)
	}

	return allSubSchedules, nil
}

func (s *paymentService) CreateSubscription(ctx context.Context, customerID, paymentMethodID, subscriptionPlanID string) *model.HTTPResponse {
	params := &stripe.SubscriptionCreateParams{
		Customer:             stripe.String(customerID),
		DefaultPaymentMethod: stripe.String(paymentMethodID),
		Items: []*stripe.SubscriptionCreateItemParams{
			{
				Price: stripe.String(subscriptionPlanID),
			},
		},
		Currency:        stripe.String("thb"),
		PaymentBehavior: stripe.String("default_incomplete"),
		Expand: []*string{
			stripe.String("latest_invoice.payments.data.payment"),
			stripe.String("latest_invoice.payments.data.invoice"),
		},
	}

	sub, err := s.stripeClient.V1Subscriptions.Create(ctx, params)
	if err != nil {
		return utils.HandleStripeError(utils.FailedToCreateMsg+"subscription: ", err)
	}

	paymentIntent, err := s.handlePaymentIntentFromInvoiceSubscription(ctx, sub.LatestInvoice.Payments.Data)
	if err != nil {
		return utils.HandleStripeError(utils.FailedToCreateMsg+"subscription: ", err)
	}

	if paymentIntent == nil {
		if sub.LatestInvoice.Status == stripe.InvoiceStatusPaid {
			return &model.HTTPResponse{
				Status: http.StatusOK,
				Data: &model.PaymentInvoice{
					ClientSecret:           "",
					PaymentIntentStatus:    string(stripe.PaymentIntentStatusSucceeded),
					SubscriptionStatus:     string(sub.Status),
					DefaultPaymentMethodID: "",
					Amount:                 sub.LatestInvoice.AmountDue / 100,
				},
			}
		} else {
			return utils.HandleStripeError(utils.FailedToCreateMsg+"subscription: ", fmt.Errorf("payment intent failed"))
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: &model.PaymentInvoice{
			ClientSecret:           paymentIntent.ClientSecret,
			PaymentIntentStatus:    string(paymentIntent.Status),
			SubscriptionStatus:     string(sub.Status),
			DefaultPaymentMethodID: sub.DefaultPaymentMethod.ID,
			Amount:                 paymentIntent.Amount / 100,
		},
	}
}

func (s *paymentService) UpdateSubscription(ctx context.Context, customerID, newSubscriptionPlanID string) *model.HTTPResponse {
	currentSub := s.stripeClient.V1Subscriptions.List(ctx, &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
		Status:   stripe.String(string(stripe.SubscriptionStatusActive)),
		Expand: []*string{
			stripe.String("data.customer.subscriptions.data"),
			stripe.String("data.items.data"),
			stripe.String("data.items.data.price"),
			stripe.String("data.latest_invoice.payments.data"),
		},
	})

	var sub *stripe.Subscription
	for cs, err := range currentSub {
		if err != nil {
			return utils.HandleStripeError("failed to check existing subscription: ", err)
		}
		if cs.Customer.Subscriptions.Data[0].Status == stripe.SubscriptionStatusActive {
			sub = cs
			break
		}
	}

	if sub == nil {
		return &model.HTTPResponse{
			Status:  http.StatusNotFound,
			Message: utils.FailedToUpdateMsg + "subscription: user has no active subscription",
		}
	}

	subUpdated, err := s.stripeClient.V1Subscriptions.Update(ctx, sub.ID, &stripe.SubscriptionUpdateParams{
		Items: []*stripe.SubscriptionUpdateItemParams{
			{
				ID:    stripe.String(sub.Items.Data[0].ID),
				Price: stripe.String(newSubscriptionPlanID),
			},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
		Expand: []*string{
			stripe.String("customer.subscriptions.data"),
			stripe.String("latest_invoice.payments.data.payment"),
		},
		BillingCycleAnchorNow: stripe.Bool(true),
	})
	if err != nil {
		return utils.HandleStripeError("failed to update subscription: ", err)
	}

	paymentIntent, err := s.handlePaymentIntentFromInvoiceSubscription(ctx, subUpdated.LatestInvoice.Payments.Data)
	if err != nil {
		return utils.HandleStripeError("failed to update subscription: ", err)
	}

	if paymentIntent == nil {
		if subUpdated.LatestInvoice.Status == stripe.InvoiceStatusPaid {
			return &model.HTTPResponse{
				Status: http.StatusOK,
				Data: &model.PaymentInvoice{
					ClientSecret:           "",
					PaymentIntentStatus:    string(stripe.PaymentIntentStatusSucceeded),
					SubscriptionStatus:     string(subUpdated.Status),
					DefaultPaymentMethodID: "",
					Amount:                 subUpdated.LatestInvoice.AmountDue / 100,
				},
			}
		} else {
			return utils.HandleStripeError("failed to update subscription: ", fmt.Errorf("payment intent failed"))
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: &model.PaymentInvoice{
			ClientSecret:           paymentIntent.ClientSecret,
			PaymentIntentStatus:    string(paymentIntent.Status),
			SubscriptionStatus:     string(subUpdated.Status),
			DefaultPaymentMethodID: subUpdated.DefaultPaymentMethod.ID,
			Amount:                 paymentIntent.Amount / 100,
		},
	}
}

func (s *paymentService) handlePaymentIntentFromInvoiceSubscription(ctx context.Context, invoice []*stripe.InvoicePayment) (*stripe.PaymentIntent, error) {
	var paymentIntentID string
	if len(invoice) == 0 {
		return nil, nil
	} else {
		paymentIntentID = invoice[0].Payment.PaymentIntent.ID
	}

	paymentIntent, err := s.stripeClient.V1PaymentIntents.Retrieve(ctx, *stripe.String(paymentIntentID), &stripe.PaymentIntentRetrieveParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to access payment intent: %w", err)
	}

	return paymentIntent, nil
}
