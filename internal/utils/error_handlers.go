package utils

import (
	"errors"
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/stripe/stripe-go/v84"
)

var (
	ErrNoRowsUpdated  = errors.New("no rows affected when updating")
	FailedToGetMsg    = "failed to get "
	NotFoundMsg       = " not found"
	FailedToCreateMsg = "failed to create "
	FailedToUpdateMsg = "failed to update "

	ErrUnauth       = errors.New("the token is invalid")
	ErrUnauthHeader = errors.New("the user is unauthorization")
	ErrFailToGet    = errors.New("failed to get data")
)

func HandleStripeCardError(err error) *model.HTTPResponse {
	var stripeErr *stripe.Error
	if errors.As(err, &stripeErr) {
		switch stripeErr.Code {
		case stripe.ErrorCodeCardDeclined:
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "card declined",
			}
		case stripe.ErrorCodeExpiredCard:
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "card expired",
			}
		case stripe.ErrorCodeCardDeclineRateLimitExceeded:
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "card decline rate limit exceeded",
			}
		case stripe.ErrorCodeCardholderPhoneNumberRequired:
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "cardholder phone number required",
			}
		default:
			return &model.HTTPResponse{
				Status:  http.StatusInternalServerError,
				Message: FailedToUpdateMsg + "payment method",
			}
		}
	}

	return &model.HTTPResponse{
		Status:  http.StatusInternalServerError,
		Message: FailedToUpdateMsg + "payment method",
	}
}
