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
	SourceList(ctx context.Context) ([]models.Source, error)
}

type Source interface {
	ID() int64
	Name() string
	IntervalFetch(ctx context.Context) ([]models.Item, error)
	FetchFromUser(ctx context.Context, userID int64, link string) (models.Item, error)
}

type Fetcher struct {
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
) *Fetcher {
	return &Fetcher{
		articles:       articleStorage,
		sources:        soureProvider,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
		log:            log,
	}
}

func (f *Fetcher) IntervalFetch(ctx context.Context) error {
	sources, err := f.sources.SourceList(ctx)
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

			if err := f.saveItems(ctx, rssSource.ID(), items); err != nil {
				f.log.Error("Processing items from source", "source name", rssSource.Name(), "err", err.Error())
				return
			}
		}(rssSource)
	}

	wg.Wait()

	return nil
}

func (f *Fetcher) FetchFromUser(ctx context.Context, userID int64, link string) {
	src := models.Source{}

	rssSource := source.NewRRSSourceFromModel(src)

	item, err := rssSource.FetchFromUser(ctx, userID, link)
	if err != nil {
		f.log.Error("Fetching items from user", "user id", userID, "err", err.Error())
		return
	}

	if err := f.saveItem(ctx, userID, item); err != nil {
		f.log.Error("Processing items from user", "user id", userID, "err", err.Error())
		return
	}
}

func (f *Fetcher) saveItem(ctx context.Context, userID int64, item models.Item) error {
	if f.itemShouldBeSkipped(item) {
		return nil
	}

	if err := f.articles.Save(ctx, models.Article{
		UserID:      userID,
		Title:       item.Title,
		Link:        item.Link,
		Excerpt:     item.Excerpt,
		ImageURL:    item.ImageURL,
		PublishedAt: item.Date,
	}); err != nil {
		return fmt.Errorf("can't save item: %v", err)
	}

	return nil
}

func (f *Fetcher) saveItems(ctx context.Context, rssSourceID int64, items []models.Item) error {
	wg := new(sync.WaitGroup)

	for _, item := range items {
		wg.Add(1)

		go func(item models.Item) error {
			defer wg.Done()

			if f.itemShouldBeSkipped(item) {
				return nil
			}

			if err := f.articles.Save(ctx, models.Article{
				SourceID:    rssSourceID,
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

func (f *Fetcher) itemShouldBeSkipped(item models.Item) bool {
	l := len(item.Categories)

	if l == 0 {
		for _, keyword := range f.filterKeywords {
			titleContainsKeyword := strings.Contains(strings.ToLower(item.Title), keyword)

			if titleContainsKeyword {
				return false
			} else {
				return true
			}
		}
	} else {
		categoriesSet := make(map[string]struct{}, l)

		for _, category := range item.Categories {
			categoriesSet[category] = struct{}{}
		}

		for _, keyword := range f.filterKeywords {
			titleContainsKeyword := strings.Contains(strings.ToLower(item.Title), keyword)

			if _, ok := categoriesSet[keyword]; ok || titleContainsKeyword {
				return false
			}
		}
	}

	return true
}
