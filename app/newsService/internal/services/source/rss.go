package source

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"newsWebApp/app/newsService/internal/models"

	"github.com/SlyMarbo/rss"
	"github.com/go-shiori/go-readability"
)

type RSSSource struct {
	sourceURL  string
	sourceID   int64
	sourceName string
}

func NewRRSSourceFromModel(m models.Source) RSSSource {
	return RSSSource{
		sourceURL:  m.FeedURL,
		sourceID:   m.ID,
		sourceName: m.Name,
	}
}

func (s RSSSource) ID() int64 {
	return s.sourceID
}

func (s RSSSource) Name() string {
	return s.sourceName
}

func (s RSSSource) URL() string {
	return s.sourceURL
}

func (s RSSSource) IntervalFetch(ctx context.Context) ([]models.Item, error) {
	const op = "services.source.interval_fetch"

	feed, err := s.loadFeed(ctx, s.sourceURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	items := make([]models.Item, 0, len(feed.Items))

	for _, rssItem := range feed.Items {
		itm := models.Item{
			Title:      rssItem.Title,
			Categories: rssItem.Categories,
			Link:       rssItem.Link,
			Date:       rssItem.Date.UTC(),
		}

		resp, err := http.Get(itm.Link)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

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

		if article.Image == "" {
			itm.ImageURL = "/static/img/empty.png"
		} else {
			itm.ImageURL = article.Image
		}

		resp.Body.Close()

		items = append(items, itm)
	}

	return items, nil
}

func (s RSSSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {
	const op = "services.source.loaded_feed"

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
