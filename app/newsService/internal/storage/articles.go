package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"newsWebApp/app/newsService/internal/models"
)

type ArticleStorage struct {
	db *sql.DB
}

func NewArticleStorage(db *sql.DB) *ArticleStorage {
	return &ArticleStorage{db: db}
}

func (s *ArticleStorage) Save(ctx context.Context, article models.Article) error {
	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO articles (source_id, user_id, title, link, excerpt, image, published_at) 
	VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`)
	if err != nil {
		return fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx,
		article.SourceID,
		article.UserID,
		article.Title,
		article.Link,
		article.Excerpt,
		article.ImageURL,
		article.PostedAt,
	); err != nil {
		return fmt.Errorf("can't insert article: %v", err)
	}

	return nil
}

func (s *ArticleStorage) NewestNotPosted(ctx context.Context) (*models.Article, error) {
	var article = new(models.Article)
	var err error

	article, err = s.notPostedFromUsers(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			article, err = s.notPostedFromBot(ctx)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, ErrNoNewArticles
				} else {
					return nil, fmt.Errorf("can't get article: %v", err)
				}
			}

			return article, nil
		} else {
			return nil, fmt.Errorf("can't get article: %v", err)
		}
	}

	return article, nil
}

func (s *ArticleStorage) MarkPosted(ctx context.Context, id int64) error {
	stmt, err := s.db.PrepareContext(ctx, "UPDATE articles SET posted_at = $1::timestamp WHERE id = $2")
	if err != nil {
		return fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, time.Now().UTC().Format(time.RFC3339), id); err != nil {
		return fmt.Errorf("can't update article in db: %v", err)
	}

	return nil
}

func (s *ArticleStorage) notPostedFromUsers(ctx context.Context) (*models.Article, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT (user_id, title, link, excerpt, image, published_at, created_at) FROM articles '
	WHERE posted_at IS NULL AND user_id > 0 
	ORDER BY published_at DESC LIMIT 1`)
	if err != nil {
		return nil, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx)

	if err := row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("can't get source: %w", err)
	}

	article := models.Article{}

	if err := row.Scan(&article.UserID,
		&article.Title,
		&article.Link,
		&article.Excerpt,
		&article.ImageURL,
		&article.PublishedAt,
		&article.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("can't scan source: %w", err)
	}

	return &article, nil
}

func (s *ArticleStorage) notPostedFromBot(ctx context.Context) (*models.Article, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT (source_id, title, link, excerpt, image, published_at, created_at) FROM articles '
	WHERE posted_at IS NULL 
	ORDER BY published_at DESC LIMIT 1`)
	if err != nil {
		return nil, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx)

	if err := row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("can't get source: %w", err)
	}

	article := models.Article{}

	if err := row.Scan(&article.UserID,
		&article.Title,
		&article.Link,
		&article.Excerpt,
		&article.ImageURL,
		&article.PublishedAt,
		&article.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("can't scan source: %w", err)
	}

	return &article, nil
}
