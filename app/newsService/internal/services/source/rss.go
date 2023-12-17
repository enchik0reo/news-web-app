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
	URL  string
	ID   int64
	Name string
}

func NewRRSSourceFromModel(m models.Source) RSSSource {
	return RSSSource{
		URL:  m.FeedURL,
		ID:   m.ID,
		Name: m.Name,
	}
}

func (s RSSSource) Fetch(ctx context.Context) ([]models.Item, error) {
	feed, err := s.loadFeed(ctx, s.URL)
	if err != nil {
		return nil, fmt.Errorf("can't load feed: %v", err)
	}

	items := make([]models.Item, 0, len(feed.Items))

	for _, rssItem := range feed.Items {
		itm := models.Item{
			Title:      rssItem.Title,
			Categories: rssItem.Categories,
			Link:       rssItem.Link,
			Date:       rssItem.Date,
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
