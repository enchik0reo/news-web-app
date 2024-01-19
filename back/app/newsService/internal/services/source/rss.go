package source

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"newsWebApp/app/newsService/internal/models"
	"newsWebApp/app/newsService/internal/services"

	"github.com/SlyMarbo/rss"
	"github.com/go-shiori/go-readability"
)

type Cacher interface {
	CacheLink(context.Context, string) error
}

type RSSSource struct {
	cacher Cacher

	sourceURL  string
	sourceID   int64
	sourceName string
}

func NewRRSSourceFromModel(cacher Cacher, m models.Source) *RSSSource {
	return &RSSSource{
		cacher:     cacher,
		sourceURL:  m.FeedURL,
		sourceID:   m.ID,
		sourceName: m.Name,
	}
}

func (s *RSSSource) ID() int64 {
	return s.sourceID
}

func (s *RSSSource) Name() string {
	return s.sourceName
}

func (s *RSSSource) URL() string {
	return s.sourceURL
}

func (s *RSSSource) IntervalLoad(ctx context.Context) ([]models.Item, error) {
	const op = "services.source.rss.interval_load"

	feed, err := s.loadFeed(ctx, s.sourceURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	items := make([]models.Item, 0, len(feed.Items))

Loop:
	for _, rssItem := range feed.Items {
		if err := s.cacher.CacheLink(ctx, rssItem.Link); err != nil {
			if errors.Is(err, services.ErrLinkExists) {
				continue Loop
			} else {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
		}

		itm := models.Item{
			Title:      rssItem.Title,
			Categories: rssItem.Categories,
			Link:       rssItem.Link,
			Date:       rssItem.Date.UTC(),
		}

		resp, err := http.Get(itm.Link)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to download %s: %v", op, itm.Link, err)
		}
		defer resp.Body.Close()

		parsedURL, err := url.Parse(itm.Link)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		article, err := readability.FromReader(resp.Body, parsedURL)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		itm.SourceName = article.SiteName
		itm.Excerpt = article.Excerpt
		itm.ImageURL = article.Image

		resp.Body.Close()

		items = append(items, itm)
	}

	return items, nil
}

func (s *RSSSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {
	const op = "services.source.rss.load_feed"

	var feedCh = make(chan *rss.Feed)
	var errCh = make(chan error)

	go func() {
		feed, err := rss.Fetch(url)
		if err != nil {
			errCh <- err
			return
		}

		feedCh <- feed
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("%s: %w", op, ctx.Err())
	case err := <-errCh:
		return nil, fmt.Errorf("%s: %w", op, err)
	case feed := <-feedCh:
		return feed, nil
	}
}
