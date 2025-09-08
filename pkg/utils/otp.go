package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

func GenerateOTP(length int) (string, error) {
	const digits = "0123456789"
	otp := make([]byte, length)
	
	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		otp[i] = digits[num.Int64()]
	}
	
	return string(otp), nil
}

func ValidatePhoneNumber(phoneNumber string) bool {
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{6,14}$`)
	return phoneRegex.MatchString(phoneNumber)
}

func NormalizePhoneNumber(phoneNumber string) string {
	return strings.TrimSpace(phoneNumber)
}
