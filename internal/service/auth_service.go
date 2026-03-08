package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/231031/wellpaw-backend/internal/applogger"
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
	CreateUser(ctx context.Context, user *model.User) *model.HTTPResponse
	LoginUser(ctx context.Context, payload *model.LoginPayload) *model.HTTPResponse
	LoginUserWithGoogle(ctx context.Context, payload *model.LoginGooglePayload) *model.HTTPResponse
	RefreshToken(ctx context.Context, refreshToken string) *model.HTTPResponse
	RequestOTP(ctx context.Context, payload model.RequestOTPPayload) *model.HTTPResponse
	ResetPassword(ctx context.Context, payload model.ResetPasswordPayload) *model.HTTPResponse
}

type authService struct {
	userRepo          repository.UserRepository
	tokenService      TokenService
	paymentService    PaymentService
	otpService        OTPService
	googleOauthConfig *oauth2.Config
}

func NewAuthService(userRepo repository.UserRepository, tokenService TokenService, paymentService PaymentService, otpService OTPService, googleOauthConfig *oauth2.Config) AuthService {
	return &authService{
		userRepo:          userRepo,
		tokenService:      tokenService,
		paymentService:    paymentService,
		otpService:        otpService,
		googleOauthConfig: googleOauthConfig,
	}
}

func (s *authService) CreateUser(ctx context.Context, user *model.User) *model.HTTPResponse {
	customer, err := s.paymentService.GetCustomerByEmail(ctx, user.Email)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to create customer by email",
		}
	}

	if customer != nil {
		return &model.HTTPResponse{
			Status:  http.StatusConflict,
			Message: "email already exists",
		}
	}

	customerID, err := s.paymentService.CreateCustomer(ctx, user)
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to create customer in stripe: %v", err), serviceLog)
	}
	user.CustomerID = customerID

	hashed, err := s.tokenService.HashPassword(user.Password)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "user",
		}
	}

	user.Password = hashed
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "user",
		}
	}

	user.Password = ""
	return &model.HTTPResponse{
		Status: http.StatusCreated,
		Data:   user,
	}
}

func (s *authService) LoginUser(ctx context.Context, payload *model.LoginPayload) *model.HTTPResponse {
	user, err := s.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusUnauthorized,
				Message: "email not found",
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "user",
		}
	}

	if user.Password == "" {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "user not register with app, please login with google",
		}
	}

	valid, err := s.tokenService.VerifyPassword(payload.Password, user.Password)
	if err != nil || !valid {
		return &model.HTTPResponse{
			Status:  http.StatusUnauthorized,
			Message: "invalid password",
		}
	}

	userAuth := &model.UserAuth{
		ID:         user.ID,
		CustomerID: user.CustomerID,
		Tier:       user.Tier,
	}
	tokenPairs, err := s.tokenService.GenerateNewPairToken(ctx, userAuth, "")
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to login",
		}
	}

	err = s.userRepo.SetCurrentSubscriptionDetail(ctx, user.ID, user.Tier, user.SubscriptionStatus)
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to set subscription detail : %s", err.Error()), serviceLog)
	}

	user.Password = ""
	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"user":  user,
			"token": tokenPairs,
		},
	}
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) *model.HTTPResponse {
	var userAuth model.UserAuth
	tokenPairs, err := s.tokenService.GenerateNewPairToken(ctx, &userAuth, refreshToken)
	if err != nil {
		if errors.Is(err, utils.ErrUnauth) {
			return &model.HTTPResponse{
				Status:  http.StatusUnauthorized,
				Message: err.Error(),
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "new token",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   tokenPairs,
	}
}

func (s *authService) LoginUserWithGoogle(ctx context.Context, payload *model.LoginGooglePayload) *model.HTTPResponse {
	token, err := s.googleOauthConfig.Exchange(ctx, payload.AuthCode)
	if err != nil {
		if rErr, ok := err.(*oauth2.RetrieveError); ok {
			if rErr.ErrorCode == "invalid_grant" {
				return &model.HTTPResponse{
					Status:  http.StatusUnauthorized,
					Message: "Token expired or invalid",
				}
			}
			if rErr.ErrorCode == "invalid_request" {
				return &model.HTTPResponse{
					Status:  http.StatusBadRequest,
					Message: "Invalid request parameters",
				}
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to login with google",
		}
	}

	userInfo, err := s.getUserInfo(ctx, token)
	if err != nil {
		return &model.HTTPResponse{
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

			customerID, err := s.paymentService.CreateCustomer(ctx, user)
			if err != nil {
				applogger.LogError(fmt.Sprintf("failed to create customer in stripe: %v", err), serviceLog)
			}

			user.CustomerID = customerID
			err = s.userRepo.CreateUser(ctx, user)
			if err != nil {
				return &model.HTTPResponse{
					Status:  http.StatusInternalServerError,
					Message: utils.FailedToCreateMsg + "user",
				}
			}
		} else {
			return &model.HTTPResponse{
				Status:  http.StatusInternalServerError,
				Message: utils.FailedToGetMsg + "user",
			}
		}
	}

	userAuth := &model.UserAuth{
		ID:         user.ID,
		CustomerID: user.CustomerID,
		Tier:       user.Tier,
	}
	tokenPairs, err := s.tokenService.GenerateNewPairToken(ctx, userAuth, "")
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to login",
		}
	}

	err = s.userRepo.SetCurrentSubscriptionDetail(ctx, user.ID, user.Tier, user.SubscriptionStatus)
	if err != nil {
		applogger.LogError(fmt.Sprintf("failed to set subscription detail : %s", err.Error()), serviceLog)
	}

	user.Password = ""
	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"user":  user,
			"token": tokenPairs,
		},
	}
}

func (s *authService) RevokeRefreshTokenWithGoogle(ctx context.Context, refreshToken string) *model.HTTPResponse {
	return &model.HTTPResponse{
		Status: http.StatusOK,
	}
}

func (s *authService) LogoutUser(ctx context.Context, refreshToken string) *model.HTTPResponse {
	return &model.HTTPResponse{
		Status: http.StatusOK,
	}
}

func (s *authService) RequestOTP(ctx context.Context, payload model.RequestOTPPayload) *model.HTTPResponse {
	_, err := s.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user not found",
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "user",
		}
	}

	if err := s.otpService.SendOTP(ctx, payload.Email); err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to send otp",
		}
	}

	return &model.HTTPResponse{
		Status:  http.StatusOK,
		Message: "otp sent successfully",
	}
}

func (s *authService) ResetPassword(ctx context.Context, payload model.ResetPasswordPayload) *model.HTTPResponse {
	if payload.Password != payload.ConfirmedPassword {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "password and confirmed password do not match",
		}
	}

	if err := s.otpService.ValidateOTP(ctx, payload.Email, payload.OTP); err != nil {
		if errors.Is(err, ErrOTPExpired) {
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "otp expired",
			}
		}

		if errors.Is(err, ErrInvalidOTP) {
			return &model.HTTPResponse{
				Status:  http.StatusUnauthorized,
				Message: "invalid otp",
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to validate otp",
		}
	}

	hashedPassword, err := s.tokenService.HashPassword(payload.Password)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "password",
		}
	}

	if err := s.userRepo.UpdatePasswordByEmail(ctx, payload.Email, hashedPassword); err != nil {
		if errors.Is(err, utils.ErrNoRowsUpdated) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "user" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "password",
		}
	}

	return &model.HTTPResponse{
		Status:  http.StatusOK,
		Message: "reset password successfully",
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
