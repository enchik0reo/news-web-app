package services

import "errors"

var (
	ErrNoPublishedArticles = errors.New("there are no published articles")
	ErrNoNewArticles       = errors.New("there are no new articles")
	ErrNoNewArticle        = errors.New("there is no new article")
	ErrNoSources           = errors.New("there are no sources")
	ErrArticleExists       = errors.New("article already exists")
	ErrArticleSkipped      = errors.New("invalid article")
	ErrLinkExists          = errors.New("link already exists")
	ErrNoOfferedArticles   = errors.New("there are no offered articles")
	ErrArticleNotAvailable = errors.New("article not available")
	ErrInvalidUrl          = errors.New("url is invalid")
)
