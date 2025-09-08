package utils

import "fmt"

// Redis key helpers for consistent key formatting
func OTPKey(phoneNumber string) string {
	return fmt.Sprintf("otp:%s", phoneNumber)
}

func RateLimitKey(phoneNumber string) string {
	return fmt.Sprintf("rate_limit:%s", phoneNumber)
}

// Generic key builder for future extensions
func BuildKey(prefix, identifier string) string {
	return fmt.Sprintf("%s:%s", prefix, identifier)
}
