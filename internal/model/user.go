package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	PhoneNumber  string         `json:"phone_number" gorm:"uniqueIndex;not null"`
	RegisteredAt time.Time      `json:"registered_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type OTP struct {
	PhoneNumber string    `json:"phone_number"`
	Code        string    `json:"code"`
	ExpiresAt   time.Time `json:"expires_at"`
	Attempts    int       `json:"attempts"`
}

type UserResponse struct {
	ID           uint      `json:"id"`
	PhoneNumber  string    `json:"phone_number"`
	RegisteredAt time.Time `json:"registered_at"`
}

type PaginatedUsersResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:           u.ID,
		PhoneNumber:  u.PhoneNumber,
		RegisteredAt: u.RegisteredAt,
	}
}
