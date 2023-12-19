package fetcher

import (
	"context"
	"errors"
	"log/slog"

	"newsWebApp/app/webService/internal/models"
	"newsWebApp/app/webService/internal/services"
	"newsWebApp/app/webService/internal/storage"
)

type NewsService interface {
	GetArticles(ctx context.Context) ([]models.Article, error)
}

type ArticlesCache interface {
	AddArticles(ctx context.Context, articles []models.Article) error
	GetArticles(ctx context.Context) ([]models.Article, error)
}

type NewsFetcher struct {
	newsService NewsService
	newsCache   ArticlesCache
	log         *slog.Logger
}

func New(newsService NewsService, newsCache ArticlesCache, log *slog.Logger) *NewsFetcher {
	return &NewsFetcher{
		newsService: newsService,
		newsCache:   newsCache,
		log:         log,
	}
}

func (f *NewsFetcher) FetchArticles(ctx context.Context) ([]models.Article, error) {
	articles, err := f.newsCache.GetArticles(ctx)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrCacheEmpty):
			f.log.Debug("Can't get articles form cache", "err", err.Error())

			articles, err = f.newsService.GetArticles(ctx)
			if err != nil {
				switch {
				case errors.Is(err, services.ErrNoPublishedArticles):
					f.log.Debug("Can't get articles form news service", "err", err.Error())
					return nil, err
				default:
					f.log.Error("Can't get articles form news service", "err", err.Error())
					return nil, err
				}
			}

			if err := f.newsCache.AddArticles(ctx, articles); err != nil {
				f.log.Error("Can't add articles to cache", "err", err.Error())
			}

			return articles, nil
		default:
			f.log.Error("Can't get articles form cache", "err", err.Error())
			return nil, err
		}
	}

	return articles, nil
}
