package models

type ContextKey string
type ContextKeyArticle string

type Article struct {
	UserName   string
	SourceName string
	Title      string
	Link       string
	Excerpt    string
	ImageURL   string
	PostedAt   string
}

type Art struct {
	Link    string
	Content string
}

type User struct {
	ID   int64
	Name string
}
