package fetcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"newsWebApp/app/apiService/internal/models"
	"newsWebApp/app/apiService/internal/services"
	"newsWebApp/app/apiService/internal/storage"
)

type NewsService interface {
	GetNewestArticle(ctx context.Context) (*models.Article, error)
	GetArticles(ctx context.Context) ([]models.Article, error)
	GetArticlesByPage(ctx context.Context, page int64) ([]models.Article, error)
}

type ArticlesCache interface {
	AddArticles(ctx context.Context, articles []models.Article) error
	AddArticle(ctx context.Context, article *models.Article) error
	GetArticlesOnPage(ctx context.Context, page int64) ([]models.Article, error)
}

type NewsFetcher struct {
	newsService NewsService
	newsCache   ArticlesCache

	refreshInterval time.Duration
	log             *slog.Logger
}

func New(newsService NewsService, newsCache ArticlesCache, refreshInterval time.Duration, log *slog.Logger) *NewsFetcher {
	return &NewsFetcher{
		newsService:     newsService,
		newsCache:       newsCache,
		refreshInterval: refreshInterval,
		log:             log,
	}
}

func (f *NewsFetcher) Start(ctx context.Context) error {
	const op = "services.fetcher.start"

	ticker := time.NewTicker(f.refreshInterval)
	defer ticker.Stop()

	if err := f.warmUp(ctx); err != nil {
		if errors.Is(err, services.ErrNoPublishedArticles) {
			f.log.Debug("Can't warm up web service cache", "err", err.Error())
		} else {
			f.log.Error("Can't warm up web service cache", "err", err.Error())
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.intervalFetch(ctx); err != nil {
				if errors.Is(err, services.ErrNoNewArticle) {
					f.log.Debug("Can't do interval fetch", "err", err.Error())
				} else {
					f.log.Error("Can't do interval fetch", "err", err.Error())
				}
			}
		}
	}
}

func (f *NewsFetcher) FetchArticlesOnPage(ctx context.Context, page int64) ([]models.Article, error) {
	articles, err := f.newsCache.GetArticlesOnPage(ctx, page)
	if err != nil {
		if errors.Is(err, storage.ErrCacheEmpty) {
			go func() {
				if err := f.warmUp(ctx); err != nil {
					if !errors.Is(err, services.ErrNoPublishedArticles) {
						f.log.Error("Can't warm up cache", "err", err.Error())
					}
				}
			}()
		} else {
			f.log.Error("Can't get articles form cache", "err", err.Error())
		}

		articles, err := f.newsService.GetArticlesByPage(ctx, page)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrNoPublishedArticles):
				f.log.Debug("Can't get articles form news service", "err", err.Error())
				return nil, err
			case errors.Is(err, context.DeadlineExceeded):
				f.log.Debug("Can't get articles form news service", "err", err.Error())
				return nil, err
			default:
				f.log.Error("Can't get articles form news service", "err", err.Error())
				return nil, err
			}
		}

		return articles, nil
	}

	return articles, nil
}

func (f *NewsFetcher) warmUp(ctx context.Context) error {
	var err error
	var articles []models.Article

Loop:
	for i := 1; i <= 3; i++ {
		articles, err = f.newsService.GetArticles(ctx)
		if err != nil {
			if errors.Is(err, services.ErrNoPublishedArticles) {
				return err
			} else {
				time.Sleep(time.Duration(i) * time.Second)
			}
		} else {
			break Loop
		}
	}

	if err != nil {
		return err
	}

	if err := f.newsCache.AddArticles(ctx, articles); err != nil {
		f.log.Error("Can't warm up cache", "err", err.Error())
	}

	return nil
}

func (f *NewsFetcher) intervalFetch(ctx context.Context) error {
	article, err := f.newsService.GetNewestArticle(ctx)
	if err != nil {
		return err
	}

	if err := f.newsCache.AddArticle(ctx, article); err != nil {
		f.log.Error("Can't save new article in cache", "err", err.Error())
	}

	return nil
}
