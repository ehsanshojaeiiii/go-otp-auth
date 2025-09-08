package utils

import (
	"context"
	"time"
)

// Context helpers for consistent timeout management
func ShortContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

func MediumContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func LongContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// For Redis operations - typically need short timeouts
func RedisContext() (context.Context, context.CancelFunc) {
	return ShortContext()
}

// For database operations - might need more time
func DBContext() (context.Context, context.CancelFunc) {
	return MediumContext()
}
