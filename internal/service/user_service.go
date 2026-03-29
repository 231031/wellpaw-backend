package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

type UserService interface {
	GetUserByID(ctx context.Context, id uint) *model.HTTPResponse
	GetUserAllInfoByID(ctx context.Context, id uint) *model.HTTPResponse
	UpdateUser(ctx context.Context, u *model.User) *model.HTTPResponse
	UpdatePaymentMethod(ctx context.Context, id uint, paymentMethodID string) *model.HTTPResponse

	GetAllSubscriptionsPlan(ctx context.Context) *model.HTTPResponse
	GetAllSubscriptionsByCustomerID(ctx context.Context, customerID, lastID string) *model.HTTPResponse
	GetPaymentIntentByID(ctx context.Context, paymentIntentID string) *model.HTTPResponse
	StartSubscription(ctx context.Context, id uint, payload model.StartSubscriptionPayload) *model.HTTPResponse
	UpdateSubscription(ctx context.Context, customerID string, payload model.UpdateSubscriptionPayload) *model.HTTPResponse
	CancelSubscription(ctx context.Context, subscriptionID string) *model.HTTPResponse

	ManageFoodNotification(ctx context.Context, id uint) *model.HTTPResponse
	ManageCalendarNotification(ctx context.Context, id uint) *model.HTTPResponse
	ManageUpdatePetNotification(ctx context.Context, id uint) *model.HTTPResponse
}

type userService struct {
	userRepo       repository.UserRepository
	paymentService PaymentService
}

func NewUserService(userRepo repository.UserRepository, paymentService PaymentService) UserService {
	return &userService{
		userRepo:       userRepo,
		paymentService: paymentService,
	}
}

func (s *userService) GetUserByID(ctx context.Context, id uint) *model.HTTPResponse {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "user",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   user,
	}
}

func (s *userService) GetUserAllInfoByID(ctx context.Context, id uint) *model.HTTPResponse {
	user, err := s.userRepo.GetUserAllInfo(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "user",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   user,
	}
}

func (s *userService) UpdateUser(ctx context.Context, user *model.User) *model.HTTPResponse {
	err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, utils.ErrNoRowsUpdated) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "user",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   user,
	}
}

func (s *userService) UpdatePaymentMethod(ctx context.Context, id uint, paymentMethodID string) *model.HTTPResponse {
	user, err := s.userRepo.GetUserIdDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "payment method",
		}
	}

	if user.CustomerID == "" {
		customerID, err := s.paymentService.CreateCustomer(ctx, user)
		if err != nil {
			return &model.HTTPResponse{
				Status:  http.StatusInternalServerError,
				Message: utils.FailedToUpdateMsg + "payment method",
			}
		}

		err = s.userRepo.UpdateCustomerID(ctx, id, customerID)
		if err != nil {
			return &model.HTTPResponse{
				Status:  http.StatusInternalServerError,
				Message: utils.FailedToUpdateMsg + "payment method",
			}
		}

		user.CustomerID = customerID
	}

	err = s.paymentService.AttachPaymentMethod(ctx, user.CustomerID, paymentMethodID)
	if err != nil {
		return utils.HandleStripeError(utils.FailedToUpdateMsg+"payment method: ", err)
	}

	err = s.userRepo.UpdatePaymentMethod(ctx, id, paymentMethodID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "payment method",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   user,
	}
}

func (s *userService) GetAllSubscriptionsPlan(ctx context.Context) *model.HTTPResponse {
	plans, err := s.paymentService.GetAllSubscriptionsPlan(ctx)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "subscription plans",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   plans,
	}
}

func (s *userService) GetAllSubscriptionsByCustomerID(ctx context.Context, customerID, lastID string) *model.HTTPResponse {
	subscriptions, lastID, err := s.paymentService.GetAllSubscriptionsByCustomerID(ctx, customerID, lastID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "subscriptions",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"subscriptions": subscriptions,
			"last_id":       lastID,
		},
	}
}

func (s *userService) GetPaymentIntentByID(ctx context.Context, paymentIntentID string) *model.HTTPResponse {
	paymentIntent, err := s.paymentService.GetPaymentIntentByID(ctx, paymentIntentID)
	if err != nil {
		return utils.HandleStripeError(utils.FailedToGetMsg+"payment intent: ", err)
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   paymentIntent,
	}
}

func (s *userService) StartSubscription(ctx context.Context, id uint, payload model.StartSubscriptionPayload) *model.HTTPResponse {
	user, err := s.userRepo.GetUserIdDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "subscription",
		}
	}

	if user.PaymentMethodID == "" {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "payment method require for subscription non-free plan",
		}
	}
	return s.paymentService.CreateSubscription(ctx, user.CustomerID, user.PaymentMethodID, payload.SubscriptionPlanID)
}

func (s *userService) UpdateSubscription(ctx context.Context, customerID string, payload model.UpdateSubscriptionPayload) *model.HTTPResponse {
	return s.paymentService.UpdateSubscription(ctx, customerID, payload.NewSubscriptionPlanID)
}

func (s *userService) CancelSubscription(ctx context.Context, subscriptionID string) *model.HTTPResponse {
	sub, err := s.paymentService.CancelSubscription(ctx, subscriptionID)
	if err != nil {
		return utils.HandleStripeError(utils.FailedToUpdateMsg+"subscription: ", err)
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   sub,
	}
}

func (s *userService) ManageFoodNotification(ctx context.Context, id uint) *model.HTTPResponse {
	user, err := s.userRepo.GetUserIdDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "notification",
		}
	}

	user.NotiFood = !user.NotiFood
	err = s.userRepo.UpdateFoodNotification(ctx, id, user.NotiFood)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "notification",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   user,
	}
}

func (s *userService) ManageCalendarNotification(ctx context.Context, id uint) *model.HTTPResponse {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "notification",
		}
	}

	user.NotiCalendars = !user.NotiCalendars
	err = s.userRepo.UpdateCalendarNotification(ctx, id, user.NotiCalendars)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "notification",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   user,
	}
}

func (s *userService) ManageUpdatePetNotification(ctx context.Context, id uint) *model.HTTPResponse {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "notification",
		}
	}

	user.NotiUpdatePet = !user.NotiUpdatePet
	err = s.userRepo.UpdatePetUpdateNotification(ctx, id, user.NotiUpdatePet)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "notification",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   user,
	}
}
