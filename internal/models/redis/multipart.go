package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/quanxiang-cloud/fileserver/internal/models"
)

type multipartRepo struct {
	c *redis.ClusterClient
}

func (m *multipartRepo) Key(path string) string {
	return fmt.Sprintf("%s:%s", redisKey, path)
}

// NewMultipartRepo NewMultipartRepo
func NewMultipartRepo(c *redis.ClusterClient) models.MultipartRepo {
	return &multipartRepo{
		c: c,
	}
}

func (m multipartRepo) Create(ctx context.Context, path string, val interface{}, expiration time.Duration) error {
	key := m.Key(path)

	return m.c.SetEX(ctx, key, val, expiration).Err()
}

func (m multipartRepo) Get(ctx context.Context, path string) (string, error) {
	key := m.Key(path)
	s, err := m.c.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return s, nil
}

func (m multipartRepo) Delete(ctx context.Context, path string) error {
	key := m.Key(path)

	return m.c.Del(ctx, key).Err()
}
