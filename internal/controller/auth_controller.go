package controller

import (
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/service"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type AuthController interface {
	CreateUser(ctx *fiber.Ctx) error
	LoginUser(ctx *fiber.Ctx) error
	LoginUserWithGoogle(ctx *fiber.Ctx) error
	RefreshToken(ctx *fiber.Ctx) error
	LogoutUser(ctx *fiber.Ctx) error
	RequestOTP(ctx *fiber.Ctx) error
	ResetPassword(ctx *fiber.Ctx) error
}

type authController struct {
	authService service.AuthService
}

func NewAuthController(authService service.AuthService) AuthController {
	return &authController{
		authService: authService,
	}
}

// @Summary Request OTP
// @Description request otp for password reset
// @tags Authentication
// @Accept application/json
// @Produce application/json
// @Param   RequestOTPPayload body model.RequestOTPPayload true "Request OTP payload"
// @Success 200 {object} model.HTTPResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 404 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /auth/otp [post]
func (c *authController) RequestOTP(ctx *fiber.Ctx) error {
	var payload model.RequestOTPPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request body",
		})
	}
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.authService.RequestOTP(ctxWithTimeOut, payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Reset Password
// @Description reset password with email and otp
// @tags Authentication
// @Accept application/json
// @Produce application/json
// @Param   ResetPasswordPayload body model.ResetPasswordPayload true "Reset password payload"
// @Success 200 {object} model.HTTPResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 401 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /auth/resetpassword [post]
func (c *authController) ResetPassword(ctx *fiber.Ctx) error {
	var payload model.ResetPasswordPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request body",
		})
	}
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.authService.ResetPassword(ctxWithTimeOut, payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Register User
// @Description register a new user
// @tags Authentication
// @Param   RegisterPayload body model.User true "Register payload"
// @Accept application/json
// @Produce application/json
// @Success 201 {object} model.User
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /auth/register [post]
func (c *authController) CreateUser(ctx *fiber.Ctx) error {
	var user model.User
	if err := ctx.BodyParser(&user); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}
	if validationResponse, err := utils.ValidateStruct(&user); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.authService.CreateUser(ctxWithTimeOut, &user)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Login User
// @Description login a user
// @tags Authentication
// @Param   LoginPayload body model.LoginPayload true "Login payload"
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.LoginResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /auth/login [post]
func (c *authController) LoginUser(ctx *fiber.Ctx) error {
	var payload model.LoginPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.authService.LoginUser(ctxWithTimeOut, &payload)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Login User With Google
// @Description login a user with google
// @tags Authentication
// @Param   LoginGooglePayload body model.LoginGooglePayload true "Login google payload"
// @Accept application/json
// @Produce application/json
// @Success 200 {object} model.LoginResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /auth/login/google [post]
func (c *authController) LoginUserWithGoogle(ctx *fiber.Ctx) error {
	var payload model.LoginGooglePayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.authService.LoginUserWithGoogle(ctxWithTimeOut, &payload)
	return ctx.Status(response.Status).JSON(response)
}

func (c *authController) RefreshToken(ctx *fiber.Ctx) error {
	var payload model.TokenPair
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.authService.RefreshToken(ctxWithTimeOut, payload.RefreshToken)
	return ctx.Status(response.Status).JSON(response)
}

// @Summary Logout User
// @Description logout user and remove auth caches from redis
// @tags Authentication
// @Security BearerAuth
// @Accept application/json
// @Produce application/json
// @Param   LogoutPayload body model.LogoutPayload true "Logout payload"
// @Success 200 {object} model.HTTPResponse
// @Failure 400 {object} model.HTTPResponse
// @Failure 500 {object} model.HTTPResponse
// @Router /auth/logout [post]
func (c *authController) LogoutUser(ctx *fiber.Ctx) error {
	userID := ctx.Locals("id").(uint)

	var payload model.LogoutPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}
	if validationResponse, err := utils.ValidateStruct(&payload); err != nil {
		return ctx.Status(validationResponse.Status).JSON(validationResponse)
	}

	ctxWithTimeOut, cancel := withTimeout(ctx.Context(), defaultTimeout)
	defer cancel()

	response := c.authService.LogoutUser(ctxWithTimeOut, userID, payload.RefreshToken)
	return ctx.Status(response.Status).JSON(response)
}
