package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/ehsanshojaei/go-otp-auth/internal/config"
	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/ehsanshojaei/go-otp-auth/internal/repository"
	"github.com/ehsanshojaei/go-otp-auth/pkg/jwt"
	"github.com/ehsanshojaei/go-otp-auth/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrInvalidOTP        = errors.New("invalid OTP")
	ErrOTPExpired       = errors.New("OTP has expired")
	ErrTooManyAttempts  = errors.New("too many OTP attempts")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrInvalidPhoneNumber = errors.New("invalid phone number format")
)

type AuthService interface {
	SendOTP(phoneNumber string) error
	VerifyOTP(phoneNumber, otpCode string) (*model.AuthResponse, error)
}

type authService struct {
	userRepo     repository.UserRepository
	otpRepo      repository.OTPRepository
	jwtManager   *jwt.JWTManager
	config       *config.Config
}

func NewAuthService(userRepo repository.UserRepository, otpRepo repository.OTPRepository, jwtManager *jwt.JWTManager, config *config.Config) AuthService {
	return &authService{
		userRepo:   userRepo,
		otpRepo:    otpRepo,
		jwtManager: jwtManager,
		config:     config,
	}
}

func (s *authService) SendOTP(phoneNumber string) error {
	phoneNumber = utils.NormalizePhoneNumber(phoneNumber)
	
	if !utils.ValidatePhoneNumber(phoneNumber) {
		return ErrInvalidPhoneNumber
	}

	// Check rate limiting
	count, err := s.otpRepo.GetRateLimitCount(phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if count >= s.config.OTP.MaxAttempts {
		return ErrRateLimitExceeded
	}

	// Generate OTP
	otpCode, err := utils.GenerateOTP(s.config.OTP.Length)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Store OTP
	if err := s.otpRepo.StoreOTP(phoneNumber, otpCode, s.config.OTP.ExpiryMinutes); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Increment rate limit
	if err := s.otpRepo.IncrementRateLimit(phoneNumber, int(s.config.OTP.RateLimitWindow.Minutes())); err != nil {
		return fmt.Errorf("failed to increment rate limit: %w", err)
	}

	// Print OTP to console (as per requirements)
	log.Printf("OTP for %s: %s", phoneNumber, otpCode)

	return nil
}

func (s *authService) VerifyOTP(phoneNumber, otpCode string) (*model.AuthResponse, error) {
	phoneNumber = utils.NormalizePhoneNumber(phoneNumber)
	
	if !utils.ValidatePhoneNumber(phoneNumber) {
		return nil, ErrInvalidPhoneNumber
	}

	// Get stored OTP
	storedOTP, err := s.otpRepo.GetOTP(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	if storedOTP == nil {
		return nil, ErrOTPExpired
	}

	// Check if too many attempts
	if storedOTP.Attempts >= s.config.OTP.MaxAttempts {
		s.otpRepo.DeleteOTP(phoneNumber)
		return nil, ErrTooManyAttempts
	}

	// Verify OTP
	if storedOTP.Code != otpCode {
		// Increment attempts
		if err := s.otpRepo.IncrementAttempts(phoneNumber); err != nil {
			log.Printf("Failed to increment OTP attempts: %v", err)
		}
		return nil, ErrInvalidOTP
	}

	// OTP is valid, delete it
	if err := s.otpRepo.DeleteOTP(phoneNumber); err != nil {
		log.Printf("Failed to delete OTP: %v", err)
	}

	// Check if user exists
	user, err := s.userRepo.GetByPhoneNumber(phoneNumber)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Create user if doesn't exist
	if user == nil {
		user = &model.User{
			PhoneNumber: phoneNumber,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.PhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &model.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}
