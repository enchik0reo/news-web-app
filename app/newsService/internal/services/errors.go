package services

import "errors"

var (
	ErrNoPublishArticles = errors.New("there are no publish articles")
	ErrNoNewArticles     = errors.New("there are no new articles")
)
