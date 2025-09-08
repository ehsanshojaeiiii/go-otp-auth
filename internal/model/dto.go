package model

import "github.com/go-playground/validator/v10"

type SendOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required" validate:"required,e164" example:"+1234567890"`
}

type VerifyOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required" validate:"required,e164" example:"+1234567890"`
	OTPCode     string `json:"otp_code" binding:"required,len=6" validate:"required,len=6" example:"123456"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type GetUsersRequest struct {
	Page        int    `form:"page" binding:"min=1" example:"1"`
	PageSize    int    `form:"page_size" binding:"min=1,max=100" example:"10"`
	PhoneNumber string `form:"phone_number" example:"+1234567890"`
}

func (r *GetUsersRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PageSize == 0 {
		r.PageSize = 10
	}
}

func (r *GetUsersRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
