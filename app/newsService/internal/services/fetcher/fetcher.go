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

type RSSSource interface {
	ID() int64
	Name() string
	IntervalLoad(ctx context.Context) ([]models.Item, error)
}

type UserSource interface {
	LoadFromUser(ctx context.Context) (models.Item, error)
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

func (f *Fetcher) Start(ctx context.Context) error {
	const op = "services.fetcher.start"

	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.intervalFetch(ctx); err != nil {
		f.log.Error("Can't do interval fetch", "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.intervalFetch(ctx); err != nil {
				f.log.Error("Can't do interval fetch", "err", err.Error())
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}
}

func (f *Fetcher) SaveArticleFromUser(ctx context.Context, userID int64, link string) error {
	const op = "services.fetcher.save_item-from_user"

	userSource := source.NewUserSource(userID, link)

	item, err := userSource.LoadFromUser(ctx)
	if err != nil {
		f.log.Error("Can't fetch items from link", "link", link, "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	if f.itemShouldBeSkipped(item) {
		return nil
	}

	if err := f.articles.Save(ctx, models.Article{
		UserID:      userID,
		SourceName:  item.SourceName,
		Title:       item.Title,
		Link:        item.Link,
		Excerpt:     item.Excerpt,
		ImageURL:    item.ImageURL,
		PublishedAt: item.Date,
	}); err != nil {
		f.log.Error("Can't save item", "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (f *Fetcher) intervalFetch(ctx context.Context) error {
	const op = "services.fetcher.interval_fetch"

	sources, err := f.sources.GetList(ctx)
	if err != nil {
		f.log.Error("Can't get soures", "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	wg := new(sync.WaitGroup)

	for _, src := range sources {
		wg.Add(1)

		rssSource := source.NewRRSSourceFromModel(src)

		go func(rssSource RSSSource) {
			defer wg.Done()

			items, err := rssSource.IntervalLoad(ctx)
			if err != nil {
				f.log.Error("Can't fetch items from source", "source name", rssSource.Name(), "err", err.Error())
				return
			}

			if err := f.saveItems(ctx, items); err != nil {
				f.log.Error("Can't save items in articles", "source name", rssSource.Name(), "err", err.Error())
				return
			}
		}(rssSource)
	}

	wg.Wait()

	return nil
}

func (f *Fetcher) saveItems(ctx context.Context, items []models.Item) error {
	const op = "services.fetcher.save_items"

	wg := new(sync.WaitGroup)

	for _, item := range items {
		wg.Add(1)

		go func(item models.Item) error {
			defer wg.Done()

			if f.itemShouldBeSkipped(item) {
				return nil
			}

			if err := f.articles.Save(ctx, models.Article{
				SourceName:  item.SourceName,
				Title:       item.Title,
				Link:        item.Link,
				Excerpt:     item.Excerpt,
				ImageURL:    item.ImageURL,
				PublishedAt: item.Date,
			}); err != nil {
				f.log.Error("Can't save item", "err", err.Error())
				return fmt.Errorf("%s: %w", op, err)
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
