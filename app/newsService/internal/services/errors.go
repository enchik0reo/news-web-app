package services

import "errors"

var (
	ErrNoLatestArticles = errors.New("there are no latest articles")
	ErrNoNewArticles    = errors.New("there are no new articles")
)
