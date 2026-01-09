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
	ManageFoodNotification(ctx context.Context, id uint) *model.HTTPResponse
	ManageCalendarNotification(ctx context.Context, id uint) *model.HTTPResponse
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
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

func (s *userService) ManageFoodNotification(ctx context.Context, id uint) *model.HTTPResponse {
	user, err := s.userRepo.GetUserAllInfo(ctx, id)
	if err != nil {
		if errors.Is(err, utils.ErrNoRowsUpdated) {
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
	user, err := s.userRepo.GetUserAllInfo(ctx, id)
	if err != nil {
		if errors.Is(err, utils.ErrNoRowsUpdated) {
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
