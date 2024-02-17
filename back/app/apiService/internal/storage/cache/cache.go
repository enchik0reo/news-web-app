package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"newsWebApp/app/apiService/internal/models"
	"newsWebApp/app/apiService/internal/storage"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	c      *redis.Client
	limit  int
	actual int
	mu     *sync.Mutex
}

func New(ctx context.Context, host, port string, limit int) (*Cache, error) {
	const op = "storage.cache.New"

	c := Cache{
		limit: limit,
		mu:    new(sync.Mutex),
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

	articleJSON, err := json.Marshal(article)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	c.mu.Lock()
	if err := c.c.Set(ctx, fmt.Sprintf("article #%d", c.actual), articleJSON, 0).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	c.actual++
	c.mu.Unlock()

	return nil
}

func (c *Cache) AddArticles(ctx context.Context, articles []models.Article) error {
	const op = "storage.cache.AddArticles"

	c.mu.Lock()
	c.actual = len(articles)
	c.mu.Unlock()

	values := make(map[string]interface{}, len(articles))

	for i, article := range articles {
		articleJSON, err := json.Marshal(article)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		values[fmt.Sprintf("article #%d", i)] = articleJSON
	}

	if err := c.c.MSet(ctx, values).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Cache) GetArticlesOnPage(ctx context.Context, page int64) ([]models.Article, error) {
	const op = "storage.cache.GetArticlesOnPage"
	limit := 0

	if c.actual == 0 {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrCacheEmpty)
	}

	amount := int(page) * c.limit
	if c.actual >= amount {
		limit = amount
	} else {
		limit = c.actual
	}

	articles := make([]models.Article, 0, limit)
	keys := make([]string, 0, limit)

	for i := c.actual - 1; len(keys) < limit; i-- {
		keys = append(keys, fmt.Sprintf("article #%d", i))
	}

	res, err := c.c.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for _, r := range res {
		re, ok := r.(string)
		var article models.Article
		if ok {
			err = json.Unmarshal([]byte(re), &article)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
		}

		articles = append(articles, article)
	}
	return articles, nil
}

func (c *Cache) CloseConn() error {
	return c.c.Close()
}
