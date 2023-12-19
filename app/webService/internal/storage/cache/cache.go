package cache

import (
	"context"
	"fmt"
	"path/filepath"
	"text/template"
	"time"

	"newsWebApp/app/webService/internal/models"
	"newsWebApp/app/webService/internal/storage"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	c         *redis.Client
	expire    time.Duration
	limit     int
	Templates map[string]*template.Template
}

func New(ctx context.Context, host, port string, expire time.Duration, limit int, tempPath string) (*Cache, error) {
	tc, err := newTemplateCache(tempPath)
	if err != nil {
		return nil, fmt.Errorf("can't create templates cache: %w", err)
	}

	c := Cache{
		expire:    expire,
		limit:     limit,
		Templates: tc,
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

func (c *Cache) AddArticles(ctx context.Context, articles []models.Article) error {
	res, err := c.c.HGetAll(ctx, "article #1").Result()
	if err != nil {
		return err
	}

	if res["Link"] != "" {
		return storage.ErrCacheNotEmpty
	}

	for i, art := range articles {
		if err := c.c.Set(ctx, fmt.Sprintf("article #%d", i+1), art, c.expire).Err(); err != nil {
			return fmt.Errorf("can't save article, %v", err)
		}
	}

	return nil
}

func (c *Cache) GetArticles(ctx context.Context) ([]models.Article, error) {
	res, err := c.c.HGetAll(ctx, "article #1").Result()
	if err != nil {
		return nil, err
	}

	if res["Link"] != "" {
		return nil, storage.ErrCacheEmpty
	}

	articles := make([]models.Article, c.limit)

	for i := range articles {
		res, err := c.c.HGetAll(ctx, fmt.Sprintf("article #%d", i+1)).Result()
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

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}