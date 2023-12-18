package notifier

import (
	"context"
	"errors"
	"fmt"
	"time"

	"newsWebApp/app/newsService/internal/models"
	"newsWebApp/app/newsService/internal/services"
	"newsWebApp/app/newsService/internal/storage"
)

type ArticleStorage interface {
	Save(ctx context.Context, article models.Article) error
	NewestNotPosted(ctx context.Context) (*models.Article, error)
	LatestPosted(ctx context.Context, limit int64) ([]models.Article, error)
	MarkPosted(ctx context.Context, id int64) error
}

type Notifier struct {
	articles      ArticleStorage
	sendInterval  time.Duration
	articlesLimit int64
}

func New(
	articles ArticleStorage,
	sendInterval time.Duration,
	articlesLimit int64,
) *Notifier {
	return &Notifier{
		articles:      articles,
		sendInterval:  sendInterval,
		articlesLimit: articlesLimit,
	}
}

func (n *Notifier) Start(ctx context.Context) error {
	ticker := time.NewTicker(n.sendInterval)
	defer ticker.Stop()

	if err := n.selectAndSendArticle(ctx); err != nil {
		return fmt.Errorf("can't select and sent article: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := n.selectAndSendArticle(ctx); err != nil {
				return fmt.Errorf("can't select and sent article in loop: %v", err)
			}
		}
	}
}

func (n *Notifier) SaveArticleFromUser(ctx context.Context, article models.Article) error {
	if err := n.articles.Save(ctx, article); err != nil {
		return fmt.Errorf("can't save article: %v", err)
	}

	return nil
}

func (n *Notifier) SelectPostedArticles(ctx context.Context, limit int64) ([]models.Article, error) {
	articles, err := n.articles.LatestPosted(ctx, limit)
	if err != nil {
		if errors.Is(err, storage.ErrNoLatestArticles) {
			return nil, services.ErrNoLatestArticles
		}
		return nil, fmt.Errorf("can't select articles: %v", err)
	}

	return articles, nil
}

func (n *Notifier) selectAndSendArticle(ctx context.Context) error {
	article, err := n.articles.NewestNotPosted(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrNoNewArticles) {
			return services.ErrNoNewArticles
		}
		return fmt.Errorf("can't select article: %v", err)
	}

	if err := n.articles.MarkPosted(ctx, article.ID); err != nil {
		return fmt.Errorf("can't mark article as posted: %v", err)
	}

	return nil
}
