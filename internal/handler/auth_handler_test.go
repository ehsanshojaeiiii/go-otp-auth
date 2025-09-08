package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/ehsanshojaei/go-otp-auth/internal/service"
	"github.com/gofiber/fiber/v2"
)

// Mock auth service for testing
type mockAuthService struct {
	sendOTPFunc   func(string) error
	verifyOTPFunc func(string, string) (*model.AuthResponse, error)
}

func (m *mockAuthService) SendOTP(phoneNumber string) error {
	if m.sendOTPFunc != nil {
		return m.sendOTPFunc(phoneNumber)
	}
	return nil
}

func (m *mockAuthService) VerifyOTP(phoneNumber, otpCode string) (*model.AuthResponse, error) {
	if m.verifyOTPFunc != nil {
		return m.verifyOTPFunc(phoneNumber, otpCode)
	}
	return &model.AuthResponse{
		Token: "test-token",
		User: model.UserResponse{
			ID:          1,
			PhoneNumber: phoneNumber,
		},
	}, nil
}

func setupTestApp() (*fiber.App, *mockAuthService) {
	mockService := &mockAuthService{}
	handler := NewAuthHandler(mockService)

	app := fiber.New()
	app.Post("/auth/send-otp", handler.SendOTP)
	app.Post("/auth/verify-otp", handler.VerifyOTP)

	return app, mockService
}

func TestAuthHandler_SendOTP(t *testing.T) {
	app, mockService := setupTestApp()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockFunc       func(string) error
		expectedStatus int
		checkResponse  bool
	}{
		{
			name: "Valid request",
			requestBody: model.SendOTPRequest{
				PhoneNumber: "+1234567890",
			},
			mockFunc:       func(string) error { return nil },
			expectedStatus: fiber.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			mockFunc:       func(string) error { return nil },
			expectedStatus: fiber.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name: "Rate limit exceeded",
			requestBody: model.SendOTPRequest{
				PhoneNumber: "+1234567890",
			},
			mockFunc:       func(string) error { return service.ErrRateLimitExceeded },
			expectedStatus: fiber.StatusTooManyRequests,
			checkResponse:  false,
		},
		{
			name: "Invalid phone number",
			requestBody: model.SendOTPRequest{
				PhoneNumber: "+1234567890",
			},
			mockFunc:       func(string) error { return service.ErrInvalidPhoneNumber },
			expectedStatus: fiber.StatusBadRequest,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.sendOTPFunc = tt.mockFunc

			var requestBody []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/auth/send-otp", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to perform request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.checkResponse {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}

				var response model.SuccessResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if response.Message == "" {
					t.Error("Expected success message, got empty")
				}
			}
		})
	}
}

func TestAuthHandler_VerifyOTP(t *testing.T) {
	app, mockService := setupTestApp()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockFunc       func(string, string) (*model.AuthResponse, error)
		expectedStatus int
		checkToken     bool
	}{
		{
			name: "Valid request",
			requestBody: model.VerifyOTPRequest{
				PhoneNumber: "+1234567890",
				OTPCode:     "123456",
			},
			mockFunc: func(string, string) (*model.AuthResponse, error) {
				return &model.AuthResponse{
					Token: "valid-token",
					User: model.UserResponse{
						ID:          1,
						PhoneNumber: "+1234567890",
					},
				}, nil
			},
			expectedStatus: fiber.StatusOK,
			checkToken:     true,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			mockFunc:       func(string, string) (*model.AuthResponse, error) { return nil, nil },
			expectedStatus: fiber.StatusBadRequest,
			checkToken:     false,
		},
		{
			name: "Invalid OTP",
			requestBody: model.VerifyOTPRequest{
				PhoneNumber: "+1234567890",
				OTPCode:     "123456",
			},
			mockFunc:       func(string, string) (*model.AuthResponse, error) { return nil, service.ErrInvalidOTP },
			expectedStatus: fiber.StatusUnauthorized,
			checkToken:     false,
		},
		{
			name: "OTP expired",
			requestBody: model.VerifyOTPRequest{
				PhoneNumber: "+1234567890",
				OTPCode:     "123456",
			},
			mockFunc:       func(string, string) (*model.AuthResponse, error) { return nil, service.ErrOTPExpired },
			expectedStatus: fiber.StatusUnauthorized,
			checkToken:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.verifyOTPFunc = tt.mockFunc

			var requestBody []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/auth/verify-otp", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to perform request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.checkToken {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}

				var response model.AuthResponse
				if err := json.Unmarshal(body, &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if response.Token == "" {
					t.Error("Expected token, got empty")
				}

				if response.User.PhoneNumber == "" {
					t.Error("Expected user data, got empty")
				}
			}
		})
	}
}
