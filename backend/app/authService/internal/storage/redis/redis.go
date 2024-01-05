package redis

import (
	"context"
	"fmt"
	"time"

	"newsWebApp/app/authService/internal/storage"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	c      *redis.Client
	expire time.Duration
}

func New(ctx context.Context, host, port string, expire time.Duration) (*Storage, error) {
	s := Storage{
		expire: expire,
	}

	addr := host + ":" + port

	s.c = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := s.c.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("can't ping to redis: %w", err)
	}

	return &s, nil
}

func (s *Storage) SetSession(ctx context.Context, userID int64, refToken string) error {
	err := s.c.Set(ctx, fmt.Sprint(userID), refToken, s.expire).Err()
	if err != nil {
		return fmt.Errorf("can't save session, %w", err)
	}

	return nil
}

func (s *Storage) GetSessionToken(ctx context.Context, userID int64) (string, error) {
	val, err := s.c.Get(ctx, fmt.Sprint(userID)).Result()
	if err == redis.Nil {
		return "", storage.ErrSessionNotFound
	} else if err != nil {
		return "", fmt.Errorf("can't get session, %w", err)
	} else {
		return val, nil
	}
}

func (s *Storage) CloseConn() error {
	return s.c.Close()
}
