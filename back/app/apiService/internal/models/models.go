package models

import (
	"time"
)

type Article struct {
	ArticleID  int64     `json:"article_id"`
	UserName   string    `json:"user_name"`
	SourceName string    `json:"source_name"`
	Title      string    `json:"title"`
	Link       string    `json:"link"`
	Excerpt    string    `json:"excerpt"`
	ImageURL   string    `json:"image_url"`
	PostedAt   time.Time `json:"posted_at"`
}

type Art struct {
	Link    string
	Content string
}
