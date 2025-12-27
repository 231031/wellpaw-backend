package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

var (
	serviceLog = "[SERVICE LOGGER]"
)

type AuthService interface {
	CreateUser(ctx context.Context, user *model.User) *utils.HTTPResponse
	LoginUser(ctx context.Context, payload *model.LoginPayload) *utils.HTTPResponse
	LoginUserWithGoogle(ctx context.Context, payload *model.LoginGooglePayload) *utils.HTTPResponse
	RefreshToken(ctx context.Context, refreshToken string) *utils.HTTPResponse
}

type authService struct {
	userRepo          repository.UserRepository
	tokenService      TokenService
	googleOauthConfig *oauth2.Config
}

func NewAuthService(userRepo repository.UserRepository, tokenService TokenService, googleOauthConfig *oauth2.Config) AuthService {
	return &authService{
		userRepo:          userRepo,
		tokenService:      tokenService,
		googleOauthConfig: googleOauthConfig,
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

	if user.Password == "" {
		return &utils.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "user not register with app, please login with google",
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

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) *utils.HTTPResponse {
	var userAuth model.UserAuth
	tokenPairs, err := s.tokenService.GenerateNewPairToken(ctx, &userAuth, refreshToken)
	if err != nil {
		if errors.Is(err, utils.ErrUnauth) {
			return &utils.HTTPResponse{
				Status:  http.StatusUnauthorized,
				Message: err.Error(),
			}
		}
		return &utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "new token",
		}
	}

	return &utils.HTTPResponse{
		Status: http.StatusOK,
		Data:   tokenPairs,
	}
}

func (s *authService) LoginUserWithGoogle(ctx context.Context, payload *model.LoginGooglePayload) *utils.HTTPResponse {
	token, err := s.googleOauthConfig.Exchange(ctx, payload.AuthCode)
	if err != nil {
		if rErr, ok := err.(*oauth2.RetrieveError); ok {
			if rErr.ErrorCode == "invalid_grant" {
				return &utils.HTTPResponse{
					Status:  http.StatusUnauthorized,
					Message: "Token expired or invalid",
				}
			}
			if rErr.ErrorCode == "invalid_request" {
				return &utils.HTTPResponse{
					Status:  http.StatusBadRequest,
					Message: "Invalid request parameters",
				}
			}
		}
		return &utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to login with google",
		}
	}

	userInfo, err := s.getUserInfo(ctx, token)
	if err != nil {
		return &utils.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to login with google",
		}
	}

	user, err := s.userRepo.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = &model.User{
				Email:       userInfo.Email,
				FirstName:   userInfo.FirstName,
				LastName:    userInfo.LastName,
				DeviceToken: payload.DeviceToken,
			}
			err = s.userRepo.CreateUser(ctx, user)
			if err != nil {
				return &utils.HTTPResponse{
					Status:  http.StatusInternalServerError,
					Message: utils.FailedToCreateMsg + "user",
				}
			}
		} else {
			return &utils.HTTPResponse{
				Status:  http.StatusInternalServerError,
				Message: utils.FailedToGetMsg + "user",
			}
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

func (s *authService) RevokeRefreshTokenWithGoogle(ctx context.Context, refreshToken string) *utils.HTTPResponse {
	return &utils.HTTPResponse{
		Status: http.StatusOK,
	}
}

func (s *authService) LogoutUser(ctx context.Context, refreshToken string) *utils.HTTPResponse {
	return &utils.HTTPResponse{
		Status: http.StatusOK,
	}
}

func (s *authService) getUserInfo(ctx context.Context, token *oauth2.Token) (*model.GoogleUserInfo, error) {
	client := s.googleOauthConfig.Client(ctx, token)

	// Make a GET request to the Google UserInfo API endpoint
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer response.Body.Close()

	// Read the response body
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	// Unmarshal the JSON response into the UserInfo struct
	var userInfo model.GoogleUserInfo
	if err := json.Unmarshal(contents, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info JSON: %w", err)
	}

	return &userInfo, nil
}
