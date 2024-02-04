package fetcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"newsWebApp/app/newsService/internal/models"
	"newsWebApp/app/newsService/internal/services"
	"newsWebApp/app/newsService/internal/services/itemHandler"
	"newsWebApp/app/newsService/internal/storage"
)

type ArticleStorage interface {
	SaveArticle(ctx context.Context, article models.Article) error
	UpdateArticle(ctx context.Context, artID int64, newArt models.Article) error
	DeleteArticle(ctx context.Context, userID int64, artID int64) error
	LinkById(ctx context.Context, artID int64) (string, error)
}

type SourceStorage interface {
	GetList(ctx context.Context) ([]models.Source, error)
}

type LinkCacher interface {
	CacheLink(ctx context.Context, link string) error
	UpdateLink(ctx context.Context, newLink string, oldLink string) error
	DeleteLink(ctx context.Context, link string) error
}

type Fetcher struct {
	articleStor ArticleStorage
	sourceStor  SourceStorage
	cacher      LinkCacher

	fetchInterval time.Duration
	keywordsSet   *sync.Map
	regularExprs  []*regexp.Regexp
	log           *slog.Logger
}

func New(
	articleStorage ArticleStorage,
	soureStorage SourceStorage,
	cacher LinkCacher,
	fetchInterval time.Duration,
	filterKeywords []string,
	log *slog.Logger,
) *Fetcher {
	l := len(filterKeywords)
	regExpr := make([]*regexp.Regexp, l)
	keySet := sync.Map{}

	for i, keyword := range filterKeywords {
		keySet.Store(keyword, struct{}{})

		r := regexp.MustCompile(fmt.Sprintf("\\b%s\\b", keyword))
		regExpr[i] = r
	}

	return &Fetcher{
		articleStor:   articleStorage,
		sourceStor:    soureStorage,
		cacher:        cacher,
		fetchInterval: fetchInterval,
		keywordsSet:   &keySet,
		regularExprs:  regExpr,
		log:           log,
	}
}

func (f *Fetcher) Start(ctx context.Context) error {
	const op = "services.fetcher.start"

	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.intervalFetch(ctx); err != nil {
		if errors.Is(err, services.ErrNoSources) {
			f.log.Debug("Can't do interval fetch", "err", err.Error())
		} else {
			f.log.Error("Can't do interval fetch", "err", err.Error())
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.intervalFetch(ctx); err != nil {
				if errors.Is(err, services.ErrNoSources) {
					f.log.Debug("Can't do interval fetch", "err", err.Error())
				} else {
					f.log.Error("Can't do interval fetch", "err", err.Error())
					return fmt.Errorf("%s: %w", op, err)
				}
			}
		}
	}
}

func (f *Fetcher) SaveArticleFromUser(ctx context.Context, userID int64, link string) error {
	const op = "services.fetcher.save_article_from_user"

	userHandler := itemHandler.NewFromUser(f.cacher, userID, link)

	item, err := userHandler.LoadItem(ctx)
	if err != nil {
		if errors.Is(err, services.ErrArticleExists) {
			f.log.Debug("Can't save article from user", "err", err.Error())
			return services.ErrArticleExists
		}
		f.log.Error("Can't load article", "link", link, "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	if f.itemShouldBeSkipped(item) {
		f.cacher.DeleteLink(ctx, link)
		return services.ErrArticleSkipped
	}

	if err := f.articleStor.SaveArticle(ctx, models.Article{
		UserID:      userID,
		SourceName:  item.SourceName,
		Title:       item.Title,
		Link:        item.Link,
		Excerpt:     item.Excerpt,
		ImageURL:    item.ImageURL,
		PublishedAt: item.Date,
	}); err != nil {
		if errors.Is(err, storage.ErrArticleExists) {
			f.log.Debug("Can't save article from user", "err", err.Error())
			return services.ErrArticleExists
		}
		f.log.Error("Can't save article from user", "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (f *Fetcher) UpdateArticleByID(ctx context.Context, userID int64, artID int64, link string) error {
	const op = "services.fetcher.update_article_by_id"

	userHandler := itemHandler.NewFromUser(f.cacher, userID, link)

	oldLink, err := f.articleStor.LinkById(ctx, artID)
	if err != nil {
		f.log.Error("Can't get old link", "Article ID", artID, "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	item, err := userHandler.UpdateItem(ctx, oldLink)
	if err != nil {
		if errors.Is(err, services.ErrArticleExists) {
			f.log.Debug("Can't update article from user", "err", err.Error())
			return services.ErrArticleExists
		}
		f.log.Error("Can't load article", "link", link, "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	if f.itemShouldBeSkipped(item) {
		f.cacher.DeleteLink(ctx, link)
		return services.ErrArticleSkipped
	}

	if err := f.articleStor.UpdateArticle(ctx, artID, models.Article{
		UserID:      userID,
		SourceName:  item.SourceName,
		Title:       item.Title,
		Link:        item.Link,
		Excerpt:     item.Excerpt,
		ImageURL:    item.ImageURL,
		PublishedAt: item.Date,
	}); err != nil {
		if errors.Is(err, storage.ErrArticleExists) {
			f.log.Debug("Can't update article from user", "err", err.Error())
			return services.ErrArticleExists
		}
		f.log.Error("Can't update article from user", "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (f *Fetcher) DeleteArticleByID(ctx context.Context, userID int64, artID int64) error {
	const op = "services.fetcher.delete_article_by_id"

	userHandler := itemHandler.NewFromUser(f.cacher, userID, "")

	oldLink, err := f.articleStor.LinkById(ctx, artID)
	if err != nil {
		f.log.Error("Can't get old link", "Article ID", artID, "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := userHandler.DeleteItem(ctx, oldLink); err != nil {
		f.log.Error("Can't delete article from cache", "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := f.articleStor.DeleteArticle(ctx, userID, artID); err != nil {
		f.log.Error("Can't delete article", "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (f *Fetcher) intervalFetch(ctx context.Context) error {
	const op = "services.fetcher.interval_fetch"

	sources, err := f.sourceStor.GetList(ctx)
	if err != nil || len(sources) == 0 {
		switch {
		case len(sources) == 0:
			return services.ErrNoSources
		case errors.Is(err, storage.ErrNoSources):
			return services.ErrNoSources
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	wg := new(sync.WaitGroup)

	for _, src := range sources {
		wg.Add(1)

		rssSource := itemHandler.NewFromRSS(f.cacher, src)

		go func(rssSource *itemHandler.RSS) {
			defer wg.Done()

			ich := make(chan models.Item)

			go func() {
				err := rssSource.IntervalLoad(ctx, f.log, ich)
				if err != nil {
					f.log.Warn("Can't fetch items from source", "source name", rssSource.SourceName(), "err", err.Error())
					return
				}
			}()

			for itm := range ich {
				if err := f.saveItem(ctx, itm); err != nil {
					f.log.Warn("Can't save items in articles", "source name", rssSource.SourceName(), "err", err.Error())
					continue
				}
			}

		}(rssSource)
	}

	wg.Wait()

	return nil
}

func (f *Fetcher) saveItem(ctx context.Context, item models.Item) error {
	const op = "services.fetcher.save_item"

	if f.itemShouldBeSkipped(item) {
		f.cacher.DeleteLink(ctx, item.Link)
		return nil
	}

	if err := f.articleStor.SaveArticle(ctx, models.Article{
		SourceName:  item.SourceName,
		Title:       item.Title,
		Link:        item.Link,
		Excerpt:     item.Excerpt,
		ImageURL:    item.ImageURL,
		PublishedAt: item.Date,
	}); err != nil {
		if !errors.Is(err, storage.ErrArticleExists) {
			f.log.Error("Can't save item", "err", err.Error())
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

func (f *Fetcher) itemShouldBeSkipped(item models.Item) bool {
	if f.categoriesContainKeyword(item.Categories) || f.titleContainsKeyword(item.Title) {
		return false
	}
	return true
}

func (f *Fetcher) categoriesContainKeyword(categories []string) bool {
	l := len(categories)

	if l == 0 {
		return false
	}

	for _, category := range categories {
		if _, categoryEqualKeyword := f.keywordsSet.Load(category); categoryEqualKeyword {
			return true
		}
	}

	return false
}

func (f *Fetcher) titleContainsKeyword(title string) bool {
	validTitle := strings.ToLower(title)

	for _, regular := range f.regularExprs {
		if regular.MatchString(validTitle) {
			return true
		}
	}

	return false
}
