package service

import (
	"context"
	"net/http"

	"github.com/231031/pethealth-backend/internal/model"
	"github.com/231031/pethealth-backend/internal/repository"
	"github.com/231031/pethealth-backend/internal/utils"
)

type AuthService interface {
	CreateUser(ctx context.Context, user *model.User) utils.HTTPResponse
	LoginUser(ctx context.Context, payload *model.LoginPayload) utils.HTTPResponse
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

func (s *authService) CreateUser(ctx context.Context, user *model.User) utils.HTTPResponse {
	err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "user",
		}
	}

	return utils.HTTPResponse{
		Status: http.StatusCreated,
		Data:   user,
	}
}

func (s *authService) LoginUser(ctx context.Context, payload *model.LoginPayload) utils.HTTPResponse {
	user, err := s.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		return utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "user",
		}
	}

	if user == nil || user.Password != payload.Password {
		return utils.HTTPResponse{
			Status:  http.StatusUnauthorized,
			Message: "invalid email or password",
		}
	}

	return utils.HTTPResponse{
		Status: http.StatusOK,
		Data:   user,
	}
}

func (s *authService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
