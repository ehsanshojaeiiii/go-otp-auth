package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	secretKey := "test-secret-key"
	expiryHours := 1
	jwtManager := NewJWTManager(secretKey, expiryHours)

	tests := []struct {
		name        string
		userID      uint
		phoneNumber string
		wantErr     bool
	}{
		{"Valid token generation", 1, "+1234567890", false},
		{"Valid with different user", 2, "+9876543210", false},
		{"Valid with zero user ID", 0, "+1111111111", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtManager.GenerateToken(tt.userID, tt.phoneNumber)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if token == "" {
					t.Error("GenerateToken() returned empty token")
				}
				
				// Verify token can be parsed
				parsedClaims, err := jwtManager.ValidateToken(token)
				if err != nil {
					t.Errorf("Generated token is invalid: %v", err)
					return
				}

				if parsedClaims.UserID != tt.userID {
					t.Errorf("Token userID = %v, want %v", parsedClaims.UserID, tt.userID)
				}

				if parsedClaims.PhoneNumber != tt.phoneNumber {
					t.Errorf("Token phoneNumber = %v, want %v", parsedClaims.PhoneNumber, tt.phoneNumber)
				}
			}
		})
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	secretKey := "test-secret-key"
	expiryHours := 1
	jwtManager := NewJWTManager(secretKey, expiryHours)

	// Generate a valid token
	userID := uint(123)
	phoneNumber := "+1234567890"
	validToken, err := jwtManager.GenerateToken(userID, phoneNumber)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Generate an expired token
	expiredClaims := Claims{
		UserID:      userID,
		PhoneNumber: phoneNumber,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, _ := expiredToken.SignedString([]byte(secretKey))

	// Generate token with wrong secret
	wrongSecretManager := NewJWTManager("wrong-secret", expiryHours)
	wrongSecretToken, _ := wrongSecretManager.GenerateToken(userID, phoneNumber)

	tests := []struct {
		name      string
		token     string
		wantErr   error
		wantClaim bool
	}{
		{"Valid token", validToken, nil, true},
		{"Empty token", "", ErrInvalidToken, false},
		{"Invalid token format", "invalid.token.format", ErrInvalidToken, false},
		{"Expired token", expiredTokenString, ErrTokenExpired, false},
		{"Wrong secret", wrongSecretToken, ErrInvalidToken, false},
		{"Malformed token", "not.a.jwt", ErrInvalidToken, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtManager.ValidateToken(tt.token)
			
			if tt.wantErr != nil {
				if err == nil || err != tt.wantErr {
					t.Errorf("ValidateToken() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateToken() unexpected error = %v", err)
				return
			}

			if tt.wantClaim {
				if claims == nil {
					t.Error("ValidateToken() returned nil claims")
					return
				}

				if claims.UserID != userID {
					t.Errorf("Claims userID = %v, want %v", claims.UserID, userID)
				}

				if claims.PhoneNumber != phoneNumber {
					t.Errorf("Claims phoneNumber = %v, want %v", claims.PhoneNumber, phoneNumber)
				}
			}
		})
	}
}

func TestJWTManager_TokenExpiry(t *testing.T) {
	secretKey := "test-secret-key"
	expiryHours := 1
	jwtManager := NewJWTManager(secretKey, expiryHours)

	token, err := jwtManager.GenerateToken(1, "+1234567890")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Check if expiry is set correctly (within 1 second tolerance)
	expectedExpiry := time.Now().Add(time.Duration(expiryHours) * time.Hour)
	actualExpiry := claims.ExpiresAt.Time
	
	if actualExpiry.Sub(expectedExpiry).Abs() > time.Second {
		t.Errorf("Token expiry mismatch. Expected around %v, got %v", expectedExpiry, actualExpiry)
	}
}
