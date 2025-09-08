package utils

import (
	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/gofiber/fiber/v2"
)

// Response helpers for cleaner handler code
func SuccessResponse(c *fiber.Ctx, message string, data ...interface{}) error {
	response := model.SuccessResponse{Message: message}
	if len(data) > 0 {
		response.Data = data[0]
	}
	return c.JSON(response)
}

func ErrorResponse(c *fiber.Ctx, code int, errorType, message string) error {
	return c.Status(code).JSON(model.ErrorResponse{
		Error:   errorType,
		Message: message,
	})
}

func BadRequest(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusBadRequest, "bad_request", message)
}

func Unauthorized(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusUnauthorized, "unauthorized", message)
}

func NotFound(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusNotFound, "not_found", message)
}

func TooManyRequests(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusTooManyRequests, "rate_limit_exceeded", message)
}

func InternalError(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusInternalServerError, "internal_error", message)
}
