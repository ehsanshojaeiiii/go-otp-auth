package service

import (
	"fmt"
	"math"

	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/ehsanshojaei/go-otp-auth/internal/repository"
)

type UserService interface {
	GetUserByID(id uint) (*model.UserResponse, error)
	GetUsers(req *model.GetUsersRequest) (*model.PaginatedUsersResponse, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUserByID(id uint) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) GetUsers(req *model.GetUsersRequest) (*model.PaginatedUsersResponse, error) {
	req.SetDefaults()

	users, total, err := s.userRepo.GetUsers(req.Page, req.PageSize, req.PhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	userResponses := make([]model.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &model.PaginatedUsersResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}
