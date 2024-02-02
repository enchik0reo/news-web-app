package redis

import (
	"context"
	"fmt"

	"newsWebApp/app/newsService/internal/storage"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	c *redis.Client
}

func New(ctx context.Context, host, port string) (*Storage, error) {
	s := Storage{}

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

func (s *Storage) SetLink(ctx context.Context, linkHash string) error {
	var value = true

	res, err := s.c.SetNX(ctx, linkHash, value, 0).Result()
	if err != nil {
		return fmt.Errorf("can't set link, %w", err)
	}

	if !res {
		return storage.ErrLinkExists
	}

	return nil
}

func (s *Storage) DeleteLink(ctx context.Context, linkHash string) error {
	res := s.c.Del(ctx, linkHash)
	if res.Err() != nil {
		return res.Err()
	}

	return nil
}

func (s *Storage) CloseConn() error {
	return s.c.Close()
}
