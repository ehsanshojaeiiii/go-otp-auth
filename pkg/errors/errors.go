package errors

import "errors"

// Common application errors - centralized for reusability
var (
	ErrInvalidOTP         = errors.New("invalid OTP")
	ErrOTPExpired        = errors.New("OTP has expired")
	ErrTooManyAttempts   = errors.New("too many OTP attempts")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrInvalidPhoneNumber = errors.New("invalid phone number format")
)
