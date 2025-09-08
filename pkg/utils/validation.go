package utils

import (
	"strings"

	apperrors "github.com/ehsanshojaei/go-otp-auth/pkg/errors"
)

// ValidateAndNormalizePhone - centralized phone validation and normalization
func ValidateAndNormalizePhone(phoneNumber string) (string, error) {
	phoneNumber = NormalizePhoneNumber(phoneNumber)
	phoneNumber = strings.TrimSpace(phoneNumber)
	
	if len(phoneNumber) > 20 || len(phoneNumber) < 8 {
		return "", apperrors.ErrInvalidPhoneNumber
	}
	
	if !ValidatePhoneNumber(phoneNumber) {
		return "", apperrors.ErrInvalidPhoneNumber
	}
	
	return phoneNumber, nil
}

// ValidateOTPCode - centralized OTP code validation
func ValidateOTPCode(otpCode string, expectedLength int) (string, error) {
	otpCode = strings.TrimSpace(otpCode)
	
	if len(otpCode) != expectedLength {
		return "", apperrors.ErrInvalidOTP
	}
	
	for _, char := range otpCode {
		if char < '0' || char > '9' {
			return "", apperrors.ErrInvalidOTP
		}
	}
	
	return otpCode, nil
}
