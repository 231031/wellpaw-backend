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
	GetCustomerByEmail(ctx context.Context, email string) (*stripe.Customer, error)
	CreateCustomer(ctx context.Context, user *model.User) (string, error)
	AttachPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) error

	GetAllSubscriptionsPlan(ctx context.Context) ([]*model.SubscriptionPlan, error)
	GetAllSubscriptionsByCustomerID(ctx context.Context, customerID string, lastID string) ([]*model.SubscriptionHistory, string, error)
	GetPaymentIntentByID(ctx context.Context, paymentIntentID string) (*model.PaymentInvoice, error)
	GetFreeTierPriceID(ctx context.Context, subscriptionPlanID string) (string, bool, error)
	CreateSubscriptionFreeTier(ctx context.Context, customerID, subscriptionPlanID string) *model.HTTPResponse
	CreateSubscription(ctx context.Context, customerID, paymentMethodID, subscriptionPlanID string) *model.HTTPResponse
	UpdateSubscription(ctx context.Context, customerID, newSubscriptionPlanID string) *model.HTTPResponse
	CancelSubscription(ctx context.Context, subscriptionID string) (*model.SubscriptionHistory, error)
}

type paymentService struct {
	stripeClient *stripe.Client
}

func NewPaymentService(stripeClient *stripe.Client) PaymentService {
	return &paymentService{
		stripeClient: stripeClient,
	}
}

func (s *paymentService) GetCustomerByEmail(ctx context.Context, email string) (*stripe.Customer, error) {
	customers := s.stripeClient.V1Customers.List(ctx, &stripe.CustomerListParams{
		Email: stripe.String(email),
	})

	for customer, err := range customers {
		if err != nil {
			return nil, err
		}
		return customer, nil
	}

	return nil, nil
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
		// Type:     stripe.String(string(stripe.PriceTypeRecurring)),
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

		var interval string
		var intervalCount int
		if sp.Recurring != nil {
			interval = *stripe.String(sp.Recurring.Interval)
			intervalCount = int(*stripe.Int64(sp.Recurring.IntervalCount))
		}
		subscriptionPlans = append(subscriptionPlans, &model.SubscriptionPlan{
			ID:            sp.ID,
			Name:          sp.Product.Name,
			Features:      features,
			Amount:        sp.UnitAmount / 100,
			Currency:      *stripe.String(sp.Currency),
			Interval:      interval,
			IntervalCount: intervalCount,
		})
	}

	return subscriptionPlans, nil
}

func (s *paymentService) GetAllSubscriptionsByCustomerID(ctx context.Context, customerID string, lastID string) ([]*model.SubscriptionHistory, string, error) {
	var newLastID string
	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
		Status:   stripe.String("all"),
		Expand: []*string{
			stripe.String("data.items.data"),
			stripe.String("data.items.data.price"),
			stripe.String("data.latest_invoice.lines.data"),
			stripe.String("data.latest_invoice.payments.data"),
		},
		ListParams: stripe.ListParams{
			Limit:   stripe.Int64(10),
			Context: ctx,
			Single:  true,
		},
	}

	if lastID != "" {
		params.StartingAfter = stripe.String(lastID)
	}
	subscriptions := s.stripeClient.V1Subscriptions.List(ctx, params)

	subscriptionHistories := []*model.SubscriptionHistory{}
	for sub, err := range subscriptions {
		if err != nil {
			return nil, "", err
		}

		history := s.mapStripeSubToSubDetails(sub)
		subscriptionHistories = append(subscriptionHistories, history)
		newLastID = sub.ID
	}

	return subscriptionHistories, newLastID, nil
}

func (s *paymentService) GetPaymentIntentByID(ctx context.Context, paymentIntentID string) (*model.PaymentInvoice, error) {
	paymentIntent, err := s.stripeClient.V1PaymentIntents.Retrieve(ctx, paymentIntentID, &stripe.PaymentIntentRetrieveParams{
		Expand: []*string{
			stripe.String("customer.subscriptions.data.default_payment_method"),
		},
	})
	if err != nil {
		return nil, err
	}

	paymentInvoice := &model.PaymentInvoice{
		ClientSecret:           paymentIntent.ClientSecret,
		PaymentIntentStatus:    string(paymentIntent.Status),
		SubscriptionStatus:     string(paymentIntent.Customer.Subscriptions.Data[0].Status),
		DefaultPaymentMethodID: string(paymentIntent.Customer.Subscriptions.Data[0].DefaultPaymentMethod.ID),
		Amount:                 paymentIntent.Amount / 100,
	}
	return paymentInvoice, nil
}

func (s *paymentService) CreateSubscriptionFreeTier(ctx context.Context, customerID, subscriptionPlanID string) *model.HTTPResponse {
	invoiceDraft, err := s.stripeClient.V1Invoices.Create(ctx, &stripe.InvoiceCreateParams{
		Customer:         stripe.String(customerID),
		Currency:         stripe.String("thb"),
		AutoAdvance:      stripe.Bool(false),
		CollectionMethod: stripe.String(string(stripe.InvoiceCollectionMethodChargeAutomatically)),
	})
	if err != nil {
		return utils.HandleStripeError(utils.FailedToCreateMsg+"subscription: ", err)
	}

	_, err = s.stripeClient.V1InvoiceItems.Create(ctx, &stripe.InvoiceItemCreateParams{
		Customer: stripe.String(customerID),
		Invoice:  stripe.String(invoiceDraft.ID),
		Pricing: &stripe.InvoiceItemCreatePricingParams{
			Price: stripe.String(subscriptionPlanID),
		},
	})
	if err != nil {
		return utils.HandleStripeError(utils.FailedToCreateMsg+"subscription: ", err)
	}

	invoiceFinalized, err := s.stripeClient.V1Invoices.FinalizeInvoice(ctx, invoiceDraft.ID, &stripe.InvoiceFinalizeInvoiceParams{})
	if err != nil {
		return utils.HandleStripeError(utils.FailedToCreateMsg+"subscription: ", err)
	}

	if invoiceFinalized.Status != stripe.InvoiceStatusPaid {
		invoiceFinalized, err = s.stripeClient.V1Invoices.Pay(ctx, invoiceFinalized.ID, &stripe.InvoicePayParams{})
		if err != nil {
			return utils.HandleStripeError(utils.FailedToCreateMsg+"subscription: ", err)
		}
	}

	if invoiceFinalized.Status != stripe.InvoiceStatusPaid {
		return utils.HandleStripeError(utils.FailedToCreateMsg+"subscription: ", fmt.Errorf("free tier invoice status is %s", invoiceFinalized.Status))
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: &model.PaymentInvoice{
			ClientSecret:           "",
			PaymentIntentStatus:    string(stripe.PaymentIntentStatusSucceeded),
			SubscriptionStatus:     string(stripe.SubscriptionStatusActive),
			DefaultPaymentMethodID: "",
			Amount:                 invoiceFinalized.AmountPaid / 100,
		},
	}
}

func (s *paymentService) GetFreeTierPriceID(ctx context.Context, subscriptionPlanID string) (string, bool, error) {
	params := &stripe.PriceListParams{
		Type:     stripe.String(string(stripe.PriceTypeOneTime)),
		Active:   stripe.Bool(true),
		Currency: stripe.String("thb"),
		Expand:   []*string{stripe.String("data.product")},
	}

	var fallbackPriceID string
	hasAnyFreeTierPrice := false
	prices := s.stripeClient.V1Prices.List(ctx, params)
	for price, err := range prices {
		if err != nil {
			return "", false, err
		}

		if price.UnitAmount != 0 {
			continue
		}
		if price.Product != nil && !price.Product.Active {
			continue
		}

		hasAnyFreeTierPrice = true
		if fallbackPriceID == "" {
			fallbackPriceID = price.ID
		}

		if price.ID == subscriptionPlanID {
			return price.ID, true, nil
		}
		if price.Product != nil && price.Product.ID == subscriptionPlanID {
			return price.ID, true, nil
		}
	}

	if subscriptionPlanID == "0" && hasAnyFreeTierPrice {
		return fallbackPriceID, true, nil
	}

	return "", false, nil
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

func (s *paymentService) CancelSubscription(ctx context.Context, subscriptionID string) (*model.SubscriptionHistory, error) {
	sub, err := s.stripeClient.V1Subscriptions.Cancel(ctx, subscriptionID, &stripe.SubscriptionCancelParams{
		Expand: []*string{
			stripe.String("items.data"),
			stripe.String("items.data.price"),
			stripe.String("latest_invoice.lines.data"),
			stripe.String("latest_invoice.payments.data"),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	subDetails := s.mapStripeSubToSubDetails(sub)
	return subDetails, nil
}

func (s *paymentService) mapStripeSubToSubDetails(sub *stripe.Subscription) *model.SubscriptionHistory {
	subDetails := &model.SubscriptionHistory{
		SubscriptionID:     sub.ID,
		SubscriptionStatus: string(sub.Status),
	}

	// --- Subscription Item ---
	if len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		price := sub.Items.Data[0].Price

		subDetails.PriceID = price.ID
		subDetails.Amount = price.UnitAmount / 100

		if price.Recurring != nil {
			subDetails.Tier = utils.ConvertIntervalToTierType(*stripe.String(price.Recurring.Interval))
		}
	}

	// --- Latest Invoice ---
	invoice := sub.LatestInvoice
	if invoice != nil {
		subDetails.InvoiceID = invoice.ID
		subDetails.InvoiceStatus = string(invoice.Status)
		subDetails.AmountPaid = invoice.AmountPaid / 100
		subDetails.AmountDue = invoice.AmountDue / 100

		// Invoice period
		if len(invoice.Lines.Data) > 0 {
			line := invoice.Lines.Data[0]
			subDetails.PeriodStart = utils.ConvertStripeTimeToTimeStr(line.Period.Start)
			subDetails.PeriodEnd = utils.ConvertStripeTimeToTimeStr(line.Period.End)
		}

		// PaymentIntent
		if len(invoice.Payments.Data) > 0 &&
			invoice.Payments.Data[0].Payment != nil &&
			invoice.Payments.Data[0].Payment.PaymentIntent != nil {

			subDetails.PaymentIntentID = invoice.Payments.Data[0].Payment.PaymentIntent.ID
		}
	}

	return subDetails
}
