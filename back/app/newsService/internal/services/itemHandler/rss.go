package itemHandler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"newsWebApp/app/newsService/internal/models"
	"newsWebApp/app/newsService/internal/services"

	"github.com/SlyMarbo/rss"
	"github.com/go-shiori/go-readability"
)

type Cacher interface {
	CacheLink(context.Context, string) error
	UpdateLink(context.Context, string, string) error
	DeleteLink(context.Context, string) error
}

type RSS struct {
	cacher Cacher

	sourceURL  string
	sourceID   int64
	sourceName string
}

func NewFromRSS(cacher Cacher, m models.Source) *RSS {
	return &RSS{
		cacher:     cacher,
		sourceURL:  m.FeedURL,
		sourceID:   m.ID,
		sourceName: m.Name,
	}
}

func (s *RSS) SourceName() string {
	return s.sourceName
}

func (s *RSS) IntervalLoad(ctx context.Context, slog *slog.Logger, ich chan models.Item) error {
	const op = "services.source.rss.interval_load"

	defer close(ich)

	feed, err := s.loadFeed(ctx, s.sourceURL)
	if err != nil {
		return fmt.Errorf("%s: failed load rss feed: %w", op, err)
	}

	wg := new(sync.WaitGroup)

	for _, rssItem := range feed.Items {
		time.Sleep(50 * time.Millisecond)
		wg.Add(1)

		go func(rssItem *rss.Item) {
			defer wg.Done()
			if err := s.cacher.CacheLink(ctx, rssItem.Link); err != nil {
				if errors.Is(err, services.ErrLinkExists) {
					return
				} else {
					slog.Debug("Can't save link in cache", "err", err.Error())
					return
				}
			}

			itm := models.Item{
				Title:      rssItem.Title,
				Categories: rssItem.Categories,
				Link:       rssItem.Link,
				Date:       rssItem.Date.UTC(),
			}

			resp, err := getResp(itm.Link)
			if err != nil {
				slog.Debug("Failed to get response", "err", err.Error())
				return
			}
			defer resp.Body.Close()

			parsedURL, err := url.Parse(itm.Link)
			if err != nil {
				slog.Debug("Failed to parse link", "err", err.Error())
				return
			}

			article, err := readability.FromReader(resp.Body, parsedURL)
			if err != nil {
				slog.Debug("Failed to parse body", "err", err.Error())
				return
			}

			if article.SiteName == "" || len([]rune(article.SiteName)) > 30 {
				article.SiteName = s.sourceName
			}

			itm.SourceName = article.SiteName
			itm.Excerpt = article.Excerpt

			if article.Image != "" {
				if article.Image[0] != 'h' {
					article.Image = ""
				}
			}

			itm.ImageURL = article.Image

			ich <- itm
		}(rssItem)
	}

	wg.Wait()

	return nil
}

func (s *RSS) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {
	var err error
	var feed *rss.Feed

	for i := 1; i <= 3; i++ {
		feed, err = s.fetch(ctx, url)
		if err != nil {
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("after retries: %w", err)
	}

	return feed, nil
}

func (s *RSS) fetch(ctx context.Context, url string) (*rss.Feed, error) {
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

func getResp(link string) (*http.Response, error) {
	var err error
	var resp *http.Response

	for i := 1; i <= 3; i++ {
		resp, err = httpGet(link)
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				time.Sleep(time.Duration(i) * time.Second)
			} else {
				return nil, err
			}
		} else {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("after retries: %w", err)
	}

	return resp, nil
}

func httpGet(url string) (*http.Response, error) {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
			DualStack: true,
		}).DialContext,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Timeout:   6 * time.Second,
		Transport: transport,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}
