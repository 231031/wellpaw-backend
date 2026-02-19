package utils

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/stripe/stripe-go/v84"
)

var (
	ErrNoRowsUpdated    = errors.New("no rows affected when updating")
	ErrFoodInActivePlan = errors.New("food is in active plan")
	FailedToGetMsg      = "failed to get "
	NotFoundMsg         = " not found"
	FailedToCreateMsg   = "failed to create "
	FailedToUpdateMsg   = "failed to update "

	ErrUnauth       = errors.New("the token is invalid")
	ErrUnauthHeader = errors.New("the user is unauthorization")
	ErrFailToGet    = errors.New("failed to get data")
)

func HandleStripeError(msgFailed string, err error) *model.HTTPResponse {
	var stripeErr *stripe.Error
	if errors.As(err, &stripeErr) {
		switch stripeErr.Type {
		case stripe.ErrorTypeCard:
			return &model.HTTPResponse{
				Status:  stripeErr.HTTPStatusCode,
				Message: msgFailed + stripeErr.Msg,
			}
		case stripe.ErrorTypeInvalidRequest:
			return &model.HTTPResponse{
				Status:  stripeErr.HTTPStatusCode,
				Message: msgFailed + stripeErr.Msg,
			}
		default:
			applogger.LogError(fmt.Sprintf("stripe error: %v", err), "STRIPE ERROR")
			return &model.HTTPResponse{
				Status:  stripeErr.HTTPStatusCode,
				Message: msgFailed + stripeErr.Msg,
			}
		}
	}

	return &model.HTTPResponse{
		Status:  http.StatusInternalServerError,
		Message: msgFailed + err.Error(),
	}
}
