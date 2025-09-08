package handler

import (
	"errors"

	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/ehsanshojaei/go-otp-auth/internal/service"
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
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	if err := h.authService.SendOTP(req.PhoneNumber); err != nil {
		switch {
		case errors.Is(err, service.ErrRateLimitExceeded):
			return c.Status(fiber.StatusTooManyRequests).JSON(model.ErrorResponse{
				Error:   "rate_limit_exceeded",
				Message: "Too many OTP requests. Please try again later.",
			})
		case errors.Is(err, service.ErrInvalidPhoneNumber):
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Error:   "invalid_phone_number",
				Message: "Phone number must be in international format (e.g., +1234567890)",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to send OTP",
			})
		}
	}

	return c.JSON(model.SuccessResponse{
		Message: "OTP sent successfully",
	})
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
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	authResponse, err := h.authService.VerifyOTP(req.PhoneNumber, req.OTPCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOTP):
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Error:   "invalid_otp",
				Message: "Invalid OTP code",
			})
		case errors.Is(err, service.ErrOTPExpired):
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Error:   "otp_expired",
				Message: "OTP has expired. Please request a new one.",
			})
		case errors.Is(err, service.ErrTooManyAttempts):
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Error:   "too_many_attempts",
				Message: "Too many failed attempts. Please request a new OTP.",
			})
		case errors.Is(err, service.ErrInvalidPhoneNumber):
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Error:   "invalid_phone_number",
				Message: "Phone number must be in international format (e.g., +1234567890)",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to verify OTP",
			})
		}
	}

	return c.JSON(authResponse)
}
