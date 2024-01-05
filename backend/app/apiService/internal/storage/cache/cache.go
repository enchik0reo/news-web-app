package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"newsWebApp/app/apiService/internal/models"
	"newsWebApp/app/apiService/internal/storage"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	c      *redis.Client
	limit  int
	actual int
}

func New(ctx context.Context, host, port string, limit int) (*Cache, error) {
	const op = "storage.cache.New"

	c := Cache{
		limit: limit,
	}

	addr := host + ":" + port

	c.c = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := c.c.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &c, nil
}

func (c *Cache) AddArticle(ctx context.Context, article *models.Article) error {
	const op = "storage.cache.AddArticle"

	articles, err := c.GetArticles(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrCacheEmpty) {
			articleJSON, err := json.Marshal(article)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			if err := c.c.Set(ctx, "article #0", articleJSON, 0).Err(); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			c.actual++
			return nil
		} else {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if c.actual < c.limit {
		articleJSON, err := json.Marshal(article)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if err := c.c.Set(ctx, "article #0", articleJSON, 0).Err(); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		for i := 0; i < c.actual; i++ {
			articleJSON, err := json.Marshal(articles[i])
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			if err := c.c.Set(ctx, fmt.Sprintf("article #%d", i+1), articleJSON, 0).Err(); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		c.actual++
	} else {
		articleJSON, err := json.Marshal(article)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if err := c.c.Set(ctx, "article #0", articleJSON, 0).Err(); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		for i := 0; i < c.limit; i++ {
			articleJSON, err := json.Marshal(articles[i])
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			if err := c.c.Set(ctx, fmt.Sprintf("article #%d", i+1), articleJSON, 0).Err(); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}

	return nil
}

func (c *Cache) AddArticles(ctx context.Context, articles []models.Article) error {
	const op = "storage.cache.AddArticles"

	c.actual = len(articles)

	for i, article := range articles {
		articleJSON, err := json.Marshal(article)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if err := c.c.Set(ctx, fmt.Sprintf("article #%d", i), articleJSON, 0).Err(); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (c *Cache) GetArticles(ctx context.Context) ([]models.Article, error) {
	const op = "storage.cache.GetArticles"

	if c.actual == 0 {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrCacheEmpty)
	}

	articles := make([]models.Article, c.actual)

	for i := range articles {
		res, err := c.c.Get(ctx, fmt.Sprintf("article #%d", i)).Result()
		if err != nil {
			if strings.Contains(err.Error(), redis.Nil.Error()) {
				return nil, fmt.Errorf("%s: %w", op, storage.ErrCacheEmpty)
			}
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		var article models.Article
		err = json.Unmarshal([]byte(res), &article)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		articles[i] = article
	}

	return articles, nil
}

func (c *Cache) CloseConn() error {
	return c.c.Close()
}
