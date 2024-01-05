package storage

import "errors"

var (
	ErrCacheEmpty    = errors.New("cache is empty")
	ErrCacheNotEmpty = errors.New("cache is not empty")
)
