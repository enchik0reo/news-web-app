package redis

import (
	"context"
	"fmt"
	"time"

	"newsWebApp/app/authService/internal/storage"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	c *redis.Client
}

func New(host, port string) (*Storage, error) {
	s := Storage{}

	addr := host + ":" + port

	s.c = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &s, nil
}

func (s *Storage) SetSession(ctx context.Context, userID int64, refToken string, expire time.Duration) error {
	err := s.c.Set(ctx, fmt.Sprint(userID), refToken, expire)
	if err != nil {
		return err.Err()
	}

	return nil
}

func (s *Storage) GetSessionToken(ctx context.Context, userID int64) (string, error) {
	val, err := s.c.Get(ctx, fmt.Sprint(userID)).Result()
	if err == redis.Nil {
		return "", storage.ErrSessionNotFound
	} else if err != nil {
		return "", fmt.Errorf("can't get session, %v", err)
	} else {
		return val, nil
	}
}

func (r *Storage) CloseConn() error {
	return r.c.Close()
}
