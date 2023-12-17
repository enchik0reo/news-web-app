package storage

import "errors"

var (
	ErrNoSources      = errors.New("there are no sources")
	ErrSourceNotFound = errors.New("source not found")
	ErrSourceExists   = errors.New("source exists")
)
