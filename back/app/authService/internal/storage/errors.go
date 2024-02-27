package storage

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user already exists")
	ErrSessionNotFound = errors.New("session not found")
	ErrNoUsers         = errors.New("there are no users")
)
