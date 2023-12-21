package cache

import (
	"context"
	"errors"
	"fmt"

	"newsWebApp/app/webService/internal/models"
	"newsWebApp/app/webService/internal/storage"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	c      *redis.Client
	limit  int
	actual int
}

func New(ctx context.Context, host, port string, limit int) (*Cache, error) {
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
		return nil, fmt.Errorf("can't ping to redis: %w", err)
	}

	return &c, nil
}

func (c *Cache) AddArticle(ctx context.Context, article *models.Article) error {
	articles, err := c.GetArticles(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrCacheEmpty) {
			if err := c.c.Set(ctx, "article #0", article, 0).Err(); err != nil {
				return fmt.Errorf("can't save article, %v", err)
			}
			c.actual++
		} else {
			return err
		}
	}

	if c.actual < c.limit {
		if err := c.c.Set(ctx, "article #0", article, 0).Err(); err != nil {
			return fmt.Errorf("can't save article, %v", err)
		}

		for i := 0; i < c.actual; i++ {
			if err := c.c.Set(ctx, fmt.Sprintf("article #%d", i+1), articles[i], 0).Err(); err != nil {
				return fmt.Errorf("can't save article, %v", err)
			}
		}

		c.actual++
	}

	return nil
}

func (c *Cache) AddArticles(ctx context.Context, articles []models.Article) error {
	c.actual = len(articles)

	for i, art := range articles {
		if err := c.c.Set(ctx, fmt.Sprintf("article #%d", i), art, 0).Err(); err != nil {
			return fmt.Errorf("can't save article, %v", err)
		}
	}

	return nil
}

func (c *Cache) GetArticles(ctx context.Context) ([]models.Article, error) {
	res, err := c.c.HGetAll(ctx, "article #0").Result()
	if err != nil {
		return nil, err
	}

	if res["Link"] != "" {
		return nil, storage.ErrCacheEmpty
	}

	articles := make([]models.Article, c.actual)

	for i := range articles {
		res, err := c.c.HGetAll(ctx, fmt.Sprintf("article #%d", i)).Result()
		if err != nil {
			return nil, err
		}

		art := models.Article{
			UserName:   res["UserName"],
			SourceName: res["SourceName"],
			Title:      res["Title"],
			Link:       res["Link"],
			Excerpt:    res["Excerpt"],
			ImageURL:   res["ImageURL"],
			PostedAt:   res["PostedAt"],
		}

		articles[i] = art
	}

	return articles, nil
}

func (c *Cache) CloseConn() error {
	return c.c.Close()
}
