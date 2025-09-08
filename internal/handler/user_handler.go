package handler

import (
	"strconv"

	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/ehsanshojaei/go-otp-auth/internal/service"
	"github.com/ehsanshojaei/go-otp-auth/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUser godoc
// @Summary Get user by ID
// @Description Retrieve a single user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} model.UserResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.BadRequest(c, "Invalid user ID format")
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.NotFound(c, "User not found")
		}
		return utils.InternalError(c, "Failed to retrieve user")
	}

	return c.JSON(user)
}

// GetUsers godoc
// @Summary Get list of users
// @Description Retrieve paginated list of users with optional search
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param phone_number query string false "Phone number search"
// @Success 200 {object} model.PaginatedUsersResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /users [get]
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	var req model.GetUsersRequest
	if err := c.QueryParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := req.Validate(); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	users, err := h.userService.GetUsers(&req)
	if err != nil {
		return utils.InternalError(c, "Failed to retrieve users")
	}

	return c.JSON(users)
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Retrieve current authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.UserResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return err
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.NotFound(c, "User not found")
		}
		return utils.InternalError(c, "Failed to retrieve profile")
	}

	return c.JSON(user)
}

// Helper method to extract user ID from JWT claims
func (h *UserHandler) getUserID(c *fiber.Ctx) (uint, error) {
	userID := c.Locals("user_id")
	if userID == nil {
		return 0, utils.Unauthorized(c, "User ID not found in token")
	}
	return userID.(uint), nil
}
