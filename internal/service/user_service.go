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
	GetUserByID(ctx context.Context, id uint) *utils.HTTPResponse
	GetUserAllInfoByID(ctx context.Context, id uint) *utils.HTTPResponse
	UpdateUser(ctx context.Context, u *model.User) *utils.HTTPResponse
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUserByID(ctx context.Context, id uint) *utils.HTTPResponse {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &utils.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}
		return &utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "user",
		}
	}

	return &utils.HTTPResponse{
		Status: http.StatusOK,
		Data:   user,
	}
}

func (s *userService) GetUserAllInfoByID(ctx context.Context, id uint) *utils.HTTPResponse {
	user, err := s.userRepo.GetUserAllInfo(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &utils.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}
		return &utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "user",
		}
	}

	return &utils.HTTPResponse{
		Status: http.StatusOK,
		Data:   user,
	}
}

func (s *userService) UpdateUser(ctx context.Context, user *model.User) *utils.HTTPResponse {
	err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, utils.ErrNoRowsUpdated) {
			return &utils.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}
		return &utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "user",
		}
	}

	return &utils.HTTPResponse{
		Status: http.StatusOK,
		Data:   user,
	}
}
