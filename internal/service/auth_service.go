package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/231031/pethealth-backend/internal/model"
	"github.com/231031/pethealth-backend/internal/repository"
	"github.com/231031/pethealth-backend/internal/utils"
	"gorm.io/gorm"
)

var (
	serviceLog = "[SERVICE LOGGER]"
)

type AuthService interface {
	CreateUser(ctx context.Context, user *model.User) *utils.HTTPResponse
	LoginUser(ctx context.Context, payload *model.LoginPayload) *utils.HTTPResponse
}

type authService struct {
	userRepo     repository.UserRepository
	tokenService TokenService
}

func NewAuthService(userRepo repository.UserRepository, tokenService TokenService) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

func (s *authService) CreateUser(ctx context.Context, user *model.User) *utils.HTTPResponse {
	hashed, err := s.tokenService.HashPassword(user.Password)
	if err != nil {
		return &utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "user",
		}
	}

	user.Password = hashed
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return &utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "user",
		}
	}

	user.Password = ""
	return &utils.HTTPResponse{
		Status: http.StatusCreated,
		Data:   user,
	}
}

func (s *authService) LoginUser(ctx context.Context, payload *model.LoginPayload) *utils.HTTPResponse {
	user, err := s.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &utils.HTTPResponse{
				Status:  http.StatusUnauthorized,
				Message: "email not found",
			}
		}

		return &utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "user",
		}
	}

	valid, err := s.tokenService.VerifyPassword(payload.Password, user.Password)
	if err != nil || !valid {
		return &utils.HTTPResponse{
			Status:  http.StatusUnauthorized,
			Message: "invalid password",
		}
	}

	userAuth := &model.UserAuth{
		ID:   user.ID,
		Tier: user.PaymentPlan,
	}
	tokenPairs, err := s.tokenService.GenerateNewPairToken(ctx, userAuth, "")
	if err != nil {
		return &utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to login",
		}
	}

	user.Password = ""
	return &utils.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"user":  user,
			"token": tokenPairs,
		},
	}
}
