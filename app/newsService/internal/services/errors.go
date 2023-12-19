package services

import "errors"

var (
	ErrNoPublishedArticles = errors.New("there are no published articles")
	ErrNoNewArticles       = errors.New("there are no new articles")
	ErrNoSources           = errors.New("there are no sources")
)
