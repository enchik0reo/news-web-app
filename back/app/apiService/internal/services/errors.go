package services

import "errors"

var (
	ErrNoPublishedArticles = errors.New("there are no published articles")
	ErrInvalidValue        = errors.New("invalid credentials")
	ErrInvalidToken        = errors.New("invalid token")
	ErrUserDoesntExists    = errors.New("user doesn't exist")
	ErrUserExists          = errors.New("user already exists")
	ErrTokenExpired        = errors.New("token expired")
	ErrSessionNotFound     = errors.New("session not found")
	ErrNoNewArticle        = errors.New("there is no new article")
	ErrArticleExists       = errors.New("article already exists")
	ErrArticleSkipped      = errors.New("invalid article")
)
