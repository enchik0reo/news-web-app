package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"newsWebApp/app/newsService/internal/models"
	"newsWebApp/app/newsService/internal/storage"

	"github.com/lib/pq"
)

type ArticleStorage struct {
	db *sql.DB
}

func NewArticleStorage(db *sql.DB) *ArticleStorage {
	return &ArticleStorage{db: db}
}

func (s *ArticleStorage) SaveArticle(ctx context.Context, article models.Article) error {
	stmt, err := s.prepareStmt(ctx, `INSERT INTO articles (user_id, source_name, title, link, excerpt, image, published_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7::timestamp)`)
	if err != nil {
		return fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	if article.UserID == 0 {
		article.UserID = 1
	}

	if err := s.retrySave(ctx, stmt, article); err != nil {
		if errors.Is(err, storage.ErrArticleExists) {
			return storage.ErrArticleExists
		} else {
			return fmt.Errorf("can't save article: %v", err)
		}
	}

	return nil
}

func (s *ArticleStorage) UpdateArticle(ctx context.Context, artID int64, article models.Article) error {
	stmt, err := s.prepareStmt(ctx, `UPDATE articles 
	SET source_name = $1, title = $2, link = $3, excerpt = $4, image = $5, created_at = $6::timestamp, published_at = $7::timestamp 
	WHERE article_id = $8 AND posted_at IS NULL RETURNING article_id`)
	if err != nil {
		return fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)

	var id int64

	if err := stmt.QueryRowContext(ctx,
		article.SourceName,
		article.Title,
		article.Link,
		article.Excerpt,
		article.ImageURL,
		now,
		article.PublishedAt,
		artID,
	).Scan(&id); err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code.Name() == "unique_violation" {
			return storage.ErrArticleExists
		} else if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrArticleNotAvailable
		} else {
			return fmt.Errorf("can't update article: %v", err)
		}
	}

	return nil
}

func (s *ArticleStorage) DeleteArticle(ctx context.Context, userID int64, artID int64) error {
	stmt, err := s.prepareStmt(ctx, `DELETE FROM articles WHERE article_id = $1 AND user_id = $2 AND posted_at IS NULL RETURNING article_id`)
	if err != nil {
		return fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	var id int64

	if err := stmt.QueryRowContext(ctx, artID, userID).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrArticleNotAvailable
		}
		return fmt.Errorf("can't get last insert id: %w", err)
	}

	return nil
}

func (s *ArticleStorage) LinkById(ctx context.Context, artID int64) (string, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT link FROM articles WHERE article_id = $1`)
	if err != nil {
		return "", fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, artID)
	if row.Err() != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrNoLink
		}
		return "", fmt.Errorf("can't get article from db: %v", err)
	}

	var link string

	if err := row.Scan(&link); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrNoLink
		}
		return "", fmt.Errorf("can't get article from db: %v", err)
	}

	return link, nil
}

func (s *ArticleStorage) ArticlesByUid(ctx context.Context, userID int64) ([]models.Article, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT article_id, u.user_name AS user_name, source_name, title, link, excerpt, image FROM articles a 
	LEFT JOIN users u ON u.user_id = a.user_id 
	WHERE a.posted_at IS NULL AND a.user_id = $1 
	ORDER BY a.created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	var articles []models.Article

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNoLatestArticles
		}
		return nil, fmt.Errorf("can't get articles from db: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		articl := models.Article{}
		err = rows.Scan(&articl.ID,
			&articl.UserName,
			&articl.SourceName,
			&articl.Title,
			&articl.Link,
			&articl.Excerpt,
			&articl.ImageURL,
		)
		if err != nil {
			return nil, fmt.Errorf("can't scan model article: %w", err)
		}

		articles = append(articles, articl)
	}

	return articles, nil
}

func (s *ArticleStorage) LatestPosted(ctx context.Context) ([]models.Article, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT article_id, u.user_name AS user_name, source_name, title, link, excerpt, image, posted_at FROM articles a 
	LEFT JOIN users u ON u.user_id = a.user_id 
	WHERE a.posted_at IS NOT NULL 
	ORDER BY a.posted_at`)
	if err != nil {
		return nil, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	var articles []models.Article

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNoLatestArticles
		}
		return nil, fmt.Errorf("can't get articles from db: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		articl := models.Article{}
		err = rows.Scan(&articl.ID,
			&articl.UserName,
			&articl.SourceName,
			&articl.Title,
			&articl.Link,
			&articl.Excerpt,
			&articl.ImageURL,
			&articl.PostedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("can't scan model article: %w", err)
		}

		articles = append(articles, articl)
	}

	return articles, nil
}

func (s *ArticleStorage) LatestPostedWithLimit(ctx context.Context, limit int64) ([]models.Article, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT article_id, u.user_name AS user_name, source_name, title, link, excerpt, image, posted_at FROM articles a 
	LEFT JOIN users u ON u.user_id = a.user_id 
	WHERE a.posted_at IS NOT NULL 
	ORDER BY a.posted_at DESC LIMIT $1`)
	if err != nil {
		return nil, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	var articles []models.Article

	rows, err := stmt.QueryContext(ctx, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNoLatestArticles
		}
		return nil, fmt.Errorf("can't get articles from db: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		articl := models.Article{}
		err = rows.Scan(&articl.ID,
			&articl.UserName,
			&articl.SourceName,
			&articl.Title,
			&articl.Link,
			&articl.Excerpt,
			&articl.ImageURL,
			&articl.PostedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("can't scan model article: %w", err)
		}

		articles = append(articles, articl)
	}

	return articles, nil
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
					return nil, storage.ErrNoNewArticles
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

func (s *ArticleStorage) MarkPosted(ctx context.Context, id int64) (time.Time, error) {
	stmt, err := s.db.PrepareContext(ctx, "UPDATE articles SET posted_at = $1::timestamp WHERE article_id = $2")
	if err != nil {
		return time.Time{}, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	postedAt := time.Now().UTC()

	if _, err := stmt.ExecContext(ctx, postedAt.Format(time.RFC3339), id); err != nil {
		return time.Time{}, fmt.Errorf("can't update article in db: %v", err)
	}

	return postedAt, nil
}

func (s *ArticleStorage) notPostedFromUsers(ctx context.Context) (*models.Article, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT u.user_name AS user_name, article_id, source_name, title, link, excerpt, image, published_at, created_at FROM articles a 
	LEFT JOIN users u ON u.user_id = a.user_id 
	WHERE a.posted_at IS NULL AND a.user_id > 1 
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

	if err := row.Scan(&article.UserName,
		&article.ID,
		&article.SourceName,
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
	stmt, err := s.db.PrepareContext(ctx, `SELECT u.user_name AS user_name, article_id, source_name, title, link, excerpt, image, published_at, created_at FROM articles a 
	LEFT JOIN users u ON u.user_id = a.user_id 
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

	if err := row.Scan(&article.UserName,
		&article.ID,
		&article.SourceName,
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

func (s *ArticleStorage) prepareStmt(ctx context.Context, query string) (*sql.Stmt, error) {
	var err error
	var stmt *sql.Stmt

	for i := 1; i <= 3; i++ {
		stmt, err = s.db.PrepareContext(ctx, query)
		if err != nil {
			pgErr, ok := err.(*pq.Error)
			if ok && pgErr.Code.Name() == "too_many_connections" {
				time.Sleep(time.Duration(i) * time.Second)
			} else {
				return nil, err
			}
		} else {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("after retries: %w", err)
	}

	return stmt, nil
}

func (s *ArticleStorage) retrySave(ctx context.Context, stmt *sql.Stmt, article models.Article) error {
	var err error

	for i := 1; i <= 3; i++ {
		if _, err := stmt.ExecContext(ctx,
			article.UserID,
			article.SourceName,
			article.Title,
			article.Link,
			article.Excerpt,
			article.ImageURL,
			article.PublishedAt.Format(time.RFC3339),
		); err != nil {
			pqErr, ok := err.(*pq.Error)
			if ok && pqErr.Code.Name() == "unique_violation" {
				return storage.ErrArticleExists
			} else {
				time.Sleep(time.Duration(i) * time.Second)
			}
		} else {
			break
		}
	}

	if err != nil {
		return err
	}

	return nil
}
