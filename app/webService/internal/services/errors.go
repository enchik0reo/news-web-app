package services

import "errors"

var (
	ErrUserDoesntExists = errors.New("user doesn't exist")
	ErrUserExists       = errors.New("user already exists")
	ErrTokenExpired     = errors.New("token expired")
	ErrSessionNotFound  = errors.New("session not found")
)
