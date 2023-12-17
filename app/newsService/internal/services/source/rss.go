package source

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

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
	feed, err := s.loadFeed(ctx, s.sourceURL)
	if err != nil {
		return nil, fmt.Errorf("can't load feed: %v", err)
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
			return nil, fmt.Errorf("failed to download %s: %v", itm.Link, err)
		}

		parsedURL, err := url.Parse(itm.Link)
		if err != nil {
			return nil, fmt.Errorf("error parsing url %s: %v", itm.Link, err)
		}

		article, err := readability.FromReader(resp.Body, parsedURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %v", itm.Link, err)
		}

		itm.SourceName = article.SiteName
		itm.ImageURL = article.Image
		itm.Excerpt = article.Excerpt

		resp.Body.Close()

		items = append(items, itm)
	}

	return items, nil
}

func (s RSSSource) FetchFromUser(ctx context.Context, userID int64, link string) (models.Item, error) {
	itm := models.Item{}

	resp, err := http.Get(link)
	if err != nil {
		return itm, fmt.Errorf("failed to download %s: %v", link, err)
	}

	parsedURL, err := url.Parse(link)
	if err != nil {
		return itm, fmt.Errorf("error parsing url %s: %v", link, err)
	}

	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return itm, fmt.Errorf("failed to parse %s: %v", link, err)
	}

	itm.Title = article.Title
	itm.Link = link
	itm.Date = time.Now().UTC()
	itm.Excerpt = article.Excerpt
	itm.ImageURL = article.Image
	itm.SourceName = article.SiteName

	resp.Body.Close()

	return itm, nil
}

func (s RSSSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {
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
		return nil, fmt.Errorf("canceled context: %v", ctx.Err())
	case err := <-errCh:
		return nil, fmt.Errorf("can't load feed rss source, %v", err)
	case feed := <-feedCh:
		return feed, nil
	}
}
