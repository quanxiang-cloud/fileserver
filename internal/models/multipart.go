package models

import (
	"context"
	"time"
)

// MultipartRepo MultipartRepo
type MultipartRepo interface {
	Create(ctx context.Context, path string, val interface{}, expiration time.Duration) error
	Get(ctx context.Context, path string) (string, error)
	Delete(ctx context.Context, path string) error
}
