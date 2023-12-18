package fetcher

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"newsWebApp/app/newsService/internal/models"
	"newsWebApp/app/newsService/internal/services/source"
)

type ArticleStorage interface {
	Save(ctx context.Context, article models.Article) error
}

type SourceStorage interface {
	GetList(ctx context.Context) ([]models.Source, error)
}

type Source interface {
	ID() int64
	Name() string
	IntervalFetch(ctx context.Context) ([]models.Item, error)
}

type IntervalFetcher struct {
	articles ArticleStorage
	sources  SourceStorage

	fetchInterval  time.Duration
	filterKeywords []string
	log            *slog.Logger
}

func New(
	articleStorage ArticleStorage,
	soureProvider SourceStorage,
	fetchInterval time.Duration,
	filterKeywords []string,
	log *slog.Logger,
) *IntervalFetcher {
	return &IntervalFetcher{
		articles:       articleStorage,
		sources:        soureProvider,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
		log:            log,
	}
}

func (f *IntervalFetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.intervalFetch(ctx); err != nil {
		return fmt.Errorf("can't do interval fetch: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.intervalFetch(ctx); err != nil {
				return fmt.Errorf("can't do interval fetch: %v", err)
			}
		}
	}
}

func (f *IntervalFetcher) intervalFetch(ctx context.Context) error {
	sources, err := f.sources.GetList(ctx)
	if err != nil {
		return fmt.Errorf("can't get soures: %v", err)
	}

	wg := new(sync.WaitGroup)

	for _, src := range sources {
		wg.Add(1)

		rssSource := source.NewRRSSourceFromModel(src)

		go func(rssSource Source) {
			defer wg.Done()

			items, err := rssSource.IntervalFetch(ctx)
			if err != nil {
				f.log.Error("Fetching items from source", "source name", rssSource.Name(), "err", err.Error())
				return
			}

			if err := f.saveItems(ctx, rssSource.Name(), items); err != nil {
				f.log.Error("Saving items in articles", "source name", rssSource.Name(), "err", err.Error())
				return
			}
		}(rssSource)
	}

	wg.Wait()

	return nil
}

func (f *IntervalFetcher) saveItems(ctx context.Context, rssSourceName string, items []models.Item) error {
	wg := new(sync.WaitGroup)

	for _, item := range items {
		wg.Add(1)

		go func(item models.Item) error {
			defer wg.Done()

			if f.itemShouldBeSkipped(item) {
				return nil
			}

			if err := f.articles.Save(ctx, models.Article{
				SourceName:  rssSourceName,
				Title:       item.Title,
				Link:        item.Link,
				Excerpt:     item.Excerpt,
				ImageURL:    item.ImageURL,
				PublishedAt: item.Date,
			}); err != nil {
				return fmt.Errorf("can't save item: %v", err)
			}

			return nil
		}(item)
	}

	wg.Wait()

	return nil
}

func (f *IntervalFetcher) itemShouldBeSkipped(item models.Item) bool {
	l := len(item.Categories)

	if l == 0 {
		for _, keyword := range f.filterKeywords {
			titleContainsKeyword := strings.Contains(strings.ToLower(item.Title), keyword)

			if titleContainsKeyword {
				return false
			}
		}
	} else {
		categoriesSet := make(map[string]struct{}, l)

		for _, category := range item.Categories {
			categoriesSet[category] = struct{}{}
		}

		for _, keyword := range f.filterKeywords {
			titleContainsKeyword := strings.Contains(strings.ToLower(item.Title), keyword)

			_, categoryContainsKeyword := categoriesSet[keyword]

			if categoryContainsKeyword || titleContainsKeyword {
				return false
			}
		}
	}

	return true
}
