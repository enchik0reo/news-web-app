package models

import "time"

type Source struct {
	ID      int64
	Name    string
	FeedURL string
}

// In RSS ...
type Item struct {
	Title      string
	Categories []string
	Link       string
	Date       time.Time
	Excerpt    string
	ImageURL   string
	SourceName string
}

// In Database ...
type Article struct {
	ID          int64
	SourceID    int64
	UserID      int64
	Title       string
	Link        string
	Excerpt     string
	ImageURL    string
	PublishedAt time.Time
	CreatedAt   time.Time
	PostedAt    time.Time
}
