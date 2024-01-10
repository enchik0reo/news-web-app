package cacher

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"

	"newsWebApp/app/newsService/internal/services"
	"newsWebApp/app/newsService/internal/storage"
)

type Setter interface {
	SetLink(context.Context, string) error
}

type Cacher struct {
	setter Setter
}

func New(setter Setter) *Cacher {
	return &Cacher{setter: setter}
}

func (c *Cacher) CacheLink(ctx context.Context, link string) error {
	linkHash, err := c.hashl(link)
	if err != nil {
		return err
	}

	if err := c.setter.SetLink(ctx, linkHash); err != nil {
		if errors.Is(err, storage.ErrLinkExists) {
			return services.ErrLinkExists
		} else {
			return err
		}
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
