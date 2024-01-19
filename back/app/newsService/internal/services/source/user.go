package source

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"newsWebApp/app/newsService/internal/models"
	"newsWebApp/app/newsService/internal/services"

	"github.com/go-shiori/go-readability"
)

type UserSource struct {
	cacher Cacher

	userID   int64
	userLink string
}

func NewUserSource(cacher Cacher, userID int64, link string) *UserSource {
	return &UserSource{
		cacher:   cacher,
		userID:   userID,
		userLink: link,
	}
}

func (s *UserSource) ID() int64 {
	return s.userID
}

func (s *UserSource) URL() string {
	return s.userLink
}

func (s *UserSource) LoadFromUser(ctx context.Context) (models.Item, error) {
	const op = "services.source.user.load_from_user"

	itm := models.Item{}

	if err := s.cacher.CacheLink(ctx, s.userLink); err != nil {
		if errors.Is(err, services.ErrLinkExists) {
			return itm, services.ErrArticleExists
		} else {
			return itm, fmt.Errorf("%s: %w", op, err)
		}
	}

	resp, err := http.Get(s.URL())
	if err != nil {
		return itm, fmt.Errorf("%s: failed to download %s: %v", op, s.URL(), err)
	}
	defer resp.Body.Close()

	parsedURL, err := url.Parse(s.URL())
	if err != nil {
		return itm, fmt.Errorf("%s: error parsing url %s: %v", op, s.URL(), err)
	}

	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return itm, fmt.Errorf("%s: failed to parse %s: %v", op, s.URL(), err)
	}

	itm.Title = article.Title
	itm.Link = s.URL()
	itm.Date = time.Now().UTC()
	itm.Excerpt = article.Excerpt
	itm.SourceName = article.SiteName

	if article.Image == "" {
		itm.ImageURL = "../img/empty.png"
	} else {
		itm.ImageURL = article.Image
	}

	resp.Body.Close()

	return itm, nil
}
