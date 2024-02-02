package cacher

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"

	"newsWebApp/app/newsService/internal/services"
	"newsWebApp/app/newsService/internal/storage"
)

type Cache interface {
	SetLink(context.Context, string) error
	DeleteLink(context.Context, string) error
}

type Cacher struct {
	cache Cache
}

func New(cache Cache) *Cacher {
	return &Cacher{cache: cache}
}

func (c *Cacher) CacheLink(ctx context.Context, link string) error {
	linkHash, err := c.hashl(link)
	if err != nil {
		return err
	}

	if err := c.cache.SetLink(ctx, linkHash); err != nil {
		if errors.Is(err, storage.ErrLinkExists) {
			return services.ErrLinkExists
		} else {
			return err
		}
	}

	return nil
}

func (c *Cacher) UpdateLink(ctx context.Context, newLink, oldLink string) error {
	newLinkHash, err := c.hashl(newLink)
	if err != nil {
		return err
	}

	oldLinkHash, err := c.hashl(oldLink)
	if err != nil {
		return err
	}

	if err := c.cache.SetLink(ctx, newLinkHash); err != nil {
		if errors.Is(err, storage.ErrLinkExists) {
			return services.ErrLinkExists
		} else {
			return err
		}
	}

	if err := c.cache.DeleteLink(ctx, oldLinkHash); err != nil {
		return err
	}

	return nil
}

func (c *Cacher) DeleteLink(ctx context.Context, link string) error {
	linkHash, err := c.hashl(link)
	if err != nil {
		return err
	}

	if err := c.cache.DeleteLink(ctx, linkHash); err != nil {
		return err
	}

	return nil
}

func (c *Cacher) hashl(link string) (string, error) {
	hash := sha1.New()

	_, err := hash.Write([]byte(link))
	if err != nil {
		return "", err
	}

	hlink := hash.Sum([]byte{})

	return fmt.Sprintf("%x", hlink), nil
}
