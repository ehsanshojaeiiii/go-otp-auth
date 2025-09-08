package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/redis/go-redis/v9"
)

type OTPRepository interface {
	StoreOTP(phoneNumber, code string, expiryMinutes int) error
	GetOTP(phoneNumber string) (*model.OTP, error)
	DeleteOTP(phoneNumber string) error
	IncrementAttempts(phoneNumber string) error
	GetRateLimitCount(phoneNumber string) (int, error)
	IncrementRateLimit(phoneNumber string, windowMinutes int) error
}

type otpRepository struct {
	client *redis.Client
}

func NewOTPRepository(client *redis.Client) OTPRepository {
	return &otpRepository{client: client}
}

func (r *otpRepository) StoreOTP(phoneNumber, code string, expiryMinutes int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	otp := model.OTP{
		PhoneNumber: phoneNumber,
		Code:        code,
		ExpiresAt:   time.Now().Add(time.Duration(expiryMinutes) * time.Minute),
		Attempts:    0,
	}

	data, err := json.Marshal(otp)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP: %w", err)
	}

	key := fmt.Sprintf("otp:%s", phoneNumber)
	return r.client.Set(ctx, key, data, time.Duration(expiryMinutes)*time.Minute).Err()
}

func (r *otpRepository) GetOTP(phoneNumber string) (*model.OTP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	key := fmt.Sprintf("otp:%s", phoneNumber)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	var otp model.OTP
	if err := json.Unmarshal([]byte(data), &otp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OTP: %w", err)
	}

	if time.Now().After(otp.ExpiresAt) {
		r.DeleteOTP(phoneNumber)
		return nil, nil
	}

	return &otp, nil
}

func (r *otpRepository) DeleteOTP(phoneNumber string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	key := fmt.Sprintf("otp:%s", phoneNumber)
	return r.client.Del(ctx, key).Err()
}

func (r *otpRepository) IncrementAttempts(phoneNumber string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	otp, err := r.GetOTP(phoneNumber)
	if err != nil || otp == nil {
		return fmt.Errorf("OTP not found")
	}

	otp.Attempts++

	data, err := json.Marshal(otp)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP: %w", err)
	}

	key := fmt.Sprintf("otp:%s", phoneNumber)
	ttl := r.client.TTL(ctx, key).Val()
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *otpRepository) GetRateLimitCount(phoneNumber string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	key := fmt.Sprintf("rate_limit:%s", phoneNumber)

	count, err := r.client.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get rate limit count: %w", err)
	}

	return count, nil
}

func (r *otpRepository) IncrementRateLimit(phoneNumber string, windowMinutes int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	key := fmt.Sprintf("rate_limit:%s", phoneNumber)

	pipe := r.client.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(windowMinutes)*time.Minute)

	_, err := pipe.Exec(ctx)
	return err
}
