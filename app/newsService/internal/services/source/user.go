package source

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"newsWebApp/app/newsService/internal/models"
	"time"

	"github.com/go-shiori/go-readability"
)

type UserSource struct {
	userID  int64
	userURL string
}

func NewUserSource(userID int64, link string) UserSource {
	return UserSource{
		userID:  userID,
		userURL: link,
	}
}

func (s UserSource) ID() int64 {
	return s.userID
}

func (s UserSource) URL() string {
	return s.userURL
}

func (s UserSource) LoadFromUser(ctx context.Context) (models.Item, error) {
	itm := models.Item{}

	resp, err := http.Get(s.URL())
	if err != nil {
		return itm, fmt.Errorf("failed to download %s: %v", s.URL(), err)
	}

	parsedURL, err := url.Parse(s.URL())
	if err != nil {
		return itm, fmt.Errorf("error parsing url %s: %v", s.URL(), err)
	}

	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return itm, fmt.Errorf("failed to parse %s: %v", s.URL(), err)
	}

	itm.Title = article.Title
	itm.Link = s.URL()
	itm.Date = time.Now().UTC()
	itm.Excerpt = article.Excerpt
	itm.SourceName = article.SiteName

	if article.Image == "" {
		itm.ImageURL = "/static/img/empty.png"
	} else {
		itm.ImageURL = article.Image
	}

	resp.Body.Close()

	return itm, nil
}
