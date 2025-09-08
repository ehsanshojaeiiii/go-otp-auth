package middleware

import (
	"strings"

	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/ehsanshojaei/go-otp-auth/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	jwtManager *jwt.JWTManager
}

func NewAuthMiddleware(jwtManager *jwt.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authorization header is required",
			})
		}

		bearerToken := strings.Fields(authHeader)
		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
			})
		}

		tokenString := bearerToken[1]
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Error:   "unauthorized",
				Message: err.Error(),
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("phone_number", claims.PhoneNumber)
		return c.Next()
	}
}
