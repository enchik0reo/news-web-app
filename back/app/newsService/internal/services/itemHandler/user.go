package itemHandler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"newsWebApp/app/newsService/internal/models"
	"newsWebApp/app/newsService/internal/services"

	"github.com/go-shiori/go-readability"
)

type User struct {
	cacher Cacher

	userID   int64
	userLink string
}

func NewFromUser(cacher Cacher, userID int64, link string) *User {
	return &User{
		cacher:   cacher,
		userID:   userID,
		userLink: link,
	}
}

func (u *User) LoadItem(ctx context.Context) (models.Item, error) {
	const op = "services.handler.user.load_from_user"

	itm := models.Item{}

	if err := u.cacher.CacheLink(ctx, u.userLink); err != nil {
		if errors.Is(err, services.ErrLinkExists) {
			return itm, services.ErrArticleExists
		} else {
			return itm, fmt.Errorf("%s: %w", op, err)
		}
	}

	resp, err := getResp(u.userLink)
	if err != nil {
		return itm, fmt.Errorf("%s: failed to download %s: %v", op, u.userLink, err)
	}
	defer resp.Body.Close()

	parsedURL, err := url.Parse(u.userLink)
	if err != nil {
		return itm, fmt.Errorf("%s: error parsing url %s: %v", op, u.userLink, err)
	}

	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return itm, fmt.Errorf("%s: failed to parse %s: %v", op, u.userLink, err)
	}

	itm.Title = article.Title
	itm.Link = u.userLink
	itm.Date = time.Now().UTC()
	itm.Excerpt = article.Excerpt
	itm.SourceName = article.SiteName

	if article.Image != "" {
		if article.Image[0] != 'h' {
			article.Image = ""
		}
	}

	itm.ImageURL = article.Image

	resp.Body.Close()

	return itm, nil
}

func (u *User) UpdateItem(ctx context.Context, oldLink string) (models.Item, error) {
	const op = "services.handler.user.update_from_user"

	itm := models.Item{}

	if err := u.cacher.UpdateLink(ctx, u.userLink, oldLink); err != nil {
		if errors.Is(err, services.ErrLinkExists) {
			return itm, services.ErrArticleExists
		} else {
			return itm, fmt.Errorf("%s: %w", op, err)
		}
	}

	resp, err := getResp(u.userLink)
	if err != nil {
		return itm, fmt.Errorf("%s: failed to download %s: %v", op, u.userLink, err)
	}
	defer resp.Body.Close()

	parsedURL, err := url.Parse(u.userLink)
	if err != nil {
		return itm, fmt.Errorf("%s: error parsing url %s: %v", op, u.userLink, err)
	}

	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return itm, fmt.Errorf("%s: failed to parse %s: %v", op, u.userLink, err)
	}

	itm.Title = article.Title
	itm.Link = u.userLink
	itm.Date = time.Now().UTC()
	itm.Excerpt = article.Excerpt
	itm.SourceName = article.SiteName

	if article.Image != "" {
		if article.Image[0] != 'h' {
			article.Image = ""
		}
	}

	itm.ImageURL = article.Image

	resp.Body.Close()

	return itm, nil
}

func (u *User) DeleteItem(ctx context.Context, link string) error {
	const op = "services.handler.user.delete_from_user"

	if err := u.cacher.DeleteLink(ctx, link); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
