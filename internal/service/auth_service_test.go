package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ehsanshojaei/go-otp-auth/internal/config"
	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/ehsanshojaei/go-otp-auth/pkg/jwt"
	"gorm.io/gorm"
)

// Mock repositories for testing
type mockUserRepository struct {
	users map[string]*model.User
	nextID uint
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*model.User),
		nextID: 1,
	}
}

func (m *mockUserRepository) Create(user *model.User) error {
	user.ID = m.nextID
	m.nextID++
	user.RegisteredAt = time.Now()
	m.users[user.PhoneNumber] = user
	return nil
}

func (m *mockUserRepository) GetByPhoneNumber(phoneNumber string) (*model.User, error) {
	user, exists := m.users[phoneNumber]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (m *mockUserRepository) GetByID(id uint) (*model.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepository) GetUsers(page, pageSize int, phoneNumber string) ([]model.User, int64, error) {
	var users []model.User
	for _, user := range m.users {
		if phoneNumber == "" || strings.Contains(user.PhoneNumber, phoneNumber) {
			users = append(users, *user)
		}
	}
	return users, int64(len(users)), nil
}

type mockOTPRepository struct {
	otps map[string]*model.OTP
	rateLimits map[string]int
}

func newMockOTPRepository() *mockOTPRepository {
	return &mockOTPRepository{
		otps: make(map[string]*model.OTP),
		rateLimits: make(map[string]int),
	}
}

func (m *mockOTPRepository) StoreOTP(phoneNumber, code string, expiryMinutes int) error {
	m.otps[phoneNumber] = &model.OTP{
		PhoneNumber: phoneNumber,
		Code:        code,
		ExpiresAt:   time.Now().Add(time.Duration(expiryMinutes) * time.Minute),
		Attempts:    0,
	}
	return nil
}

func (m *mockOTPRepository) GetOTP(phoneNumber string) (*model.OTP, error) {
	otp, exists := m.otps[phoneNumber]
	if !exists {
		return nil, nil
	}
	if time.Now().After(otp.ExpiresAt) {
		delete(m.otps, phoneNumber)
		return nil, nil
	}
	return otp, nil
}

func (m *mockOTPRepository) DeleteOTP(phoneNumber string) error {
	delete(m.otps, phoneNumber)
	return nil
}

func (m *mockOTPRepository) IncrementAttempts(phoneNumber string) error {
	otp, exists := m.otps[phoneNumber]
	if !exists {
		return errors.New("OTP not found")
	}
	otp.Attempts++
	return nil
}

func (m *mockOTPRepository) GetRateLimitCount(phoneNumber string) (int, error) {
	count, exists := m.rateLimits[phoneNumber]
	if !exists {
		return 0, nil
	}
	return count, nil
}

func (m *mockOTPRepository) IncrementRateLimit(phoneNumber string, windowMinutes int) error {
	m.rateLimits[phoneNumber]++
	return nil
}

func createTestAuthService() (AuthService, *mockUserRepository, *mockOTPRepository) {
	userRepo := newMockUserRepository()
	otpRepo := newMockOTPRepository()
	jwtManager := jwt.NewJWTManager("test-secret", 24)
	
	cfg := &config.Config{
		OTP: config.OTPConfig{
			Length:          6,
			ExpiryMinutes:   2,
			MaxAttempts:     3,
			RateLimitWindow: 10 * time.Minute,
		},
	}

	authService := NewAuthService(userRepo, otpRepo, jwtManager, cfg)
	return authService, userRepo, otpRepo
}

func TestAuthService_SendOTP(t *testing.T) {
	authService, _, otpRepo := createTestAuthService()

	tests := []struct {
		name        string
		phoneNumber string
		setupFunc   func()
		wantErr     error
	}{
		{
			name:        "Valid phone number",
			phoneNumber: "+1234567890",
			setupFunc:   func() {},
			wantErr:     nil,
		},
		{
			name:        "Invalid phone number",
			phoneNumber: "1234567890",
			setupFunc:   func() {},
			wantErr:     ErrInvalidPhoneNumber,
		},
		{
			name:        "Rate limit exceeded",
			phoneNumber: "+1111111111",
			setupFunc: func() {
				otpRepo.rateLimits["+1111111111"] = 3
			},
			wantErr: ErrRateLimitExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()
			
			err := authService.SendOTP(tt.phoneNumber)
			
			if tt.wantErr != nil {
				if err == nil || !errors.Is(err, tt.wantErr) {
					t.Errorf("SendOTP() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("SendOTP() unexpected error = %v", err)
				return
			}

			// Verify OTP was stored
			otp, err := otpRepo.GetOTP(tt.phoneNumber)
			if err != nil {
				t.Errorf("Failed to get stored OTP: %v", err)
				return
			}
			if otp == nil {
				t.Error("OTP was not stored")
				return
			}
			if len(otp.Code) != 6 {
				t.Errorf("OTP length = %v, want 6", len(otp.Code))
			}
		})
	}
}

func TestAuthService_VerifyOTP(t *testing.T) {
	authService, userRepo, otpRepo := createTestAuthService()

	// Setup: Create a valid OTP
	validPhone := "+1234567890"
	validOTP := "123456"
	otpRepo.StoreOTP(validPhone, validOTP, 2)

	// Setup: Create OTP for invalid code test
	invalidCodePhone := "+1111111112"
	invalidCodeOTP := "999999"
	otpRepo.StoreOTP(invalidCodePhone, invalidCodeOTP, 2)

	// Setup: Create an expired OTP
	expiredPhone := "+9999999999"
	expiredOTP := "654321"
	otpRepo.otps[expiredPhone] = &model.OTP{
		PhoneNumber: expiredPhone,
		Code:        expiredOTP,
		ExpiresAt:   time.Now().Add(-1 * time.Minute), // Already expired
		Attempts:    0,
	}

	// Setup: Create OTP with max attempts
	maxAttemptsPhone := "+8888888888"
	maxAttemptsOTP := "111111"
	otpRepo.otps[maxAttemptsPhone] = &model.OTP{
		PhoneNumber: maxAttemptsPhone,
		Code:        maxAttemptsOTP,
		ExpiresAt:   time.Now().Add(2 * time.Minute),
		Attempts:    3,
	}

	tests := []struct {
		name        string
		phoneNumber string
		otpCode     string
		wantErr     error
		checkResult bool
	}{
		{
			name:        "Valid OTP - new user",
			phoneNumber: validPhone,
			otpCode:     validOTP,
			wantErr:     nil,
			checkResult: true,
		},
		{
			name:        "Invalid phone format",
			phoneNumber: "1234567890",
			otpCode:     "123456",
			wantErr:     ErrInvalidPhoneNumber,
			checkResult: false,
		},
		{
			name:        "Invalid OTP code",
			phoneNumber: invalidCodePhone,
			otpCode:     "wrong",
			wantErr:     ErrInvalidOTP,
			checkResult: false,
		},
		{
			name:        "Expired OTP",
			phoneNumber: expiredPhone,
			otpCode:     expiredOTP,
			wantErr:     ErrOTPExpired,
			checkResult: false,
		},
		{
			name:        "Too many attempts",
			phoneNumber: maxAttemptsPhone,
			otpCode:     maxAttemptsOTP,
			wantErr:     ErrTooManyAttempts,
			checkResult: false,
		},
		{
			name:        "OTP not found",
			phoneNumber: "+7777777777",
			otpCode:     "123456",
			wantErr:     ErrOTPExpired,
			checkResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := authService.VerifyOTP(tt.phoneNumber, tt.otpCode)
			
			if tt.wantErr != nil {
				if err == nil || !errors.Is(err, tt.wantErr) {
					t.Errorf("VerifyOTP() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("VerifyOTP() unexpected error = %v", err)
				return
			}

			if tt.checkResult {
				if result == nil {
					t.Error("VerifyOTP() returned nil result")
					return
				}

				if result.Token == "" {
					t.Error("VerifyOTP() returned empty token")
				}

				if result.User.PhoneNumber != tt.phoneNumber {
					t.Errorf("User phone number = %v, want %v", result.User.PhoneNumber, tt.phoneNumber)
				}

				// Verify user was created
				user, err := userRepo.GetByPhoneNumber(tt.phoneNumber)
				if err != nil {
					t.Errorf("User was not created: %v", err)
				}
				if user.PhoneNumber != tt.phoneNumber {
					t.Errorf("Created user phone = %v, want %v", user.PhoneNumber, tt.phoneNumber)
				}
			}
		})
	}
}

func TestAuthService_VerifyOTP_ExistingUser(t *testing.T) {
	authService, userRepo, otpRepo := createTestAuthService()

	// Create existing user
	existingPhone := "+5555555555"
	existingUser := &model.User{
		PhoneNumber: existingPhone,
	}
	userRepo.Create(existingUser)

	// Create valid OTP
	validOTP := "123456"
	otpRepo.StoreOTP(existingPhone, validOTP, 2)

	result, err := authService.VerifyOTP(existingPhone, validOTP)
	if err != nil {
		t.Errorf("VerifyOTP() error = %v", err)
		return
	}

	if result.User.ID != existingUser.ID {
		t.Errorf("Returned user ID = %v, want %v", result.User.ID, existingUser.ID)
	}
}
