package service

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log"

	"github.com/ehsanshojaei/go-otp-auth/internal/config"
	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/ehsanshojaei/go-otp-auth/internal/repository"
	apperrors "github.com/ehsanshojaei/go-otp-auth/pkg/errors"
	"github.com/ehsanshojaei/go-otp-auth/pkg/jwt"
	"github.com/ehsanshojaei/go-otp-auth/pkg/utils"
	"gorm.io/gorm"
)

// Re-export errors for backward compatibility
var (
	ErrInvalidOTP         = apperrors.ErrInvalidOTP
	ErrOTPExpired        = apperrors.ErrOTPExpired
	ErrTooManyAttempts   = apperrors.ErrTooManyAttempts
	ErrRateLimitExceeded = apperrors.ErrRateLimitExceeded
	ErrInvalidPhoneNumber = apperrors.ErrInvalidPhoneNumber
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
	phoneNumber, err := utils.ValidateAndNormalizePhone(phoneNumber)
	if err != nil {
		return err
	}

	// Check rate limiting
	count, err := s.otpRepo.GetRateLimitCount(phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if count >= s.config.OTP.MaxAttempts {
		return ErrRateLimitExceeded
	}

	// Generate and store OTP
	otpCode, err := utils.GenerateOTP(s.config.OTP.Length)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err := s.otpRepo.StoreOTP(phoneNumber, otpCode, s.config.OTP.ExpiryMinutes); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	if err := s.otpRepo.IncrementRateLimit(phoneNumber, int(s.config.OTP.RateLimitWindow.Minutes())); err != nil {
		return fmt.Errorf("failed to increment rate limit: %w", err)
	}

	utils.LogOTP(phoneNumber, otpCode)
	return nil
}

func (s *authService) VerifyOTP(phoneNumber, otpCode string) (*model.AuthResponse, error) {
	var err error
	phoneNumber, err = utils.ValidateAndNormalizePhone(phoneNumber)
	if err != nil {
		return nil, err
	}
	
	otpCode, err = utils.ValidateOTPCode(otpCode, s.config.OTP.Length)
	if err != nil {
		return nil, err
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

	// Verify OTP using constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(storedOTP.Code), []byte(otpCode)) != 1 {
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

	// Get or create user
	user, err := s.userRepo.GetByPhoneNumber(phoneNumber)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		user = &model.User{PhoneNumber: phoneNumber}
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
