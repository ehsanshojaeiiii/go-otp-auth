package utils

import "log"

// LogOTP - centralized OTP logging for console output (per requirements)
func LogOTP(phoneNumber, otpCode string) {
	log.Printf("OTP for %s: %s", phoneNumber, otpCode)
}
