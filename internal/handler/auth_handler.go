package handler

import (
	"errors"

	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/ehsanshojaei/go-otp-auth/internal/service"
	"github.com/ehsanshojaei/go-otp-auth/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// SendOTP godoc
// @Summary Send OTP to phone number
// @Description Generate and send OTP to the provided phone number
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.SendOTPRequest true "Phone number"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 429 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /auth/send-otp [post]
func (h *AuthHandler) SendOTP(c *fiber.Ctx) error {
	var req model.SendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	err := h.authService.SendOTP(req.PhoneNumber)
	return h.handleAuthError(c, err, "OTP sent successfully")
}

// VerifyOTP godoc
// @Summary Verify OTP and login/register
// @Description Verify OTP code and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.VerifyOTPRequest true "Phone number and OTP"
// @Success 200 {object} model.AuthResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req model.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	authResponse, err := h.authService.VerifyOTP(req.PhoneNumber, req.OTPCode)
	if err != nil {
		return h.handleAuthError(c, err, "")
	}

	return c.JSON(authResponse)
}

// Helper method for consistent auth error handling
func (h *AuthHandler) handleAuthError(c *fiber.Ctx, err error, successMessage string) error {
	if err == nil {
		return utils.SuccessResponse(c, successMessage)
	}

	switch {
	case errors.Is(err, service.ErrRateLimitExceeded):
		return utils.TooManyRequests(c, "Too many OTP requests. Please try again later.")
	case errors.Is(err, service.ErrInvalidPhoneNumber):
		return utils.BadRequest(c, "Phone number must be in international format (e.g., +1234567890)")
	case errors.Is(err, service.ErrInvalidOTP):
		return utils.Unauthorized(c, "Invalid OTP code")
	case errors.Is(err, service.ErrOTPExpired):
		return utils.Unauthorized(c, "OTP has expired. Please request a new one.")
	case errors.Is(err, service.ErrTooManyAttempts):
		return utils.Unauthorized(c, "Too many failed attempts. Please request a new OTP.")
	default:
		return utils.InternalError(c, "Operation failed")
	}
}
