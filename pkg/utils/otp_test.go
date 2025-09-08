package utils

import (
	"testing"
)

func TestGenerateOTP(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   bool
	}{
		{"Valid 6 digit OTP", 6, true},
		{"Valid 4 digit OTP", 4, true},
		{"Valid 8 digit OTP", 8, true},
		{"Zero length", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp, err := GenerateOTP(tt.length)
			
			if (err == nil) != tt.want {
				t.Errorf("GenerateOTP() error = %v, want %v", err, tt.want)
				return
			}

			if err == nil {
				if len(otp) != tt.length {
					t.Errorf("GenerateOTP() length = %v, want %v", len(otp), tt.length)
				}

				// Check if all characters are digits
				for _, char := range otp {
					if char < '0' || char > '9' {
						t.Errorf("GenerateOTP() contains non-digit character: %c", char)
					}
				}
			}
		})
	}
}

func TestValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		want        bool
	}{
		{"Valid US number", "+1234567890", true},
		{"Valid international", "+4912345678901", true},
		{"Valid with country code", "+9198765432100", true},
		{"Invalid without plus", "1234567890", false},
		{"Invalid with letters", "+123abc7890", false},
		{"Invalid too short", "+123", false},
		{"Invalid too long", "+123456789012345678", false},
		{"Invalid empty", "", false},
		{"Invalid just plus", "+", false},
		{"Invalid starting with zero", "+0234567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePhoneNumber(tt.phoneNumber); got != tt.want {
				t.Errorf("ValidatePhoneNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizePhoneNumber(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		want        string
	}{
		{"No spaces", "+1234567890", "+1234567890"},
		{"Leading spaces", "  +1234567890", "+1234567890"},
		{"Trailing spaces", "+1234567890  ", "+1234567890"},
		{"Both spaces", "  +1234567890  ", "+1234567890"},
		{"Multiple spaces", "   +1234567890   ", "+1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizePhoneNumber(tt.phoneNumber); got != tt.want {
				t.Errorf("NormalizePhoneNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}
