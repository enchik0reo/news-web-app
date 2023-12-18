package notifier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	log           *slog.Logger
}

func New(
	articles ArticleStorage,
	sendInterval time.Duration,
	articlesLimit int64,
	log *slog.Logger,
) *Notifier {
	return &Notifier{
		articles:      articles,
		sendInterval:  sendInterval,
		articlesLimit: articlesLimit,
		log:           log,
	}
}

func (n *Notifier) Start(ctx context.Context) error {
	const op = "services.notifier.start"

	ticker := time.NewTicker(n.sendInterval)
	defer ticker.Stop()

	if err := n.selectAndSendArticle(ctx); err != nil {
		n.log.Error("Can't select and sent article", "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := n.selectAndSendArticle(ctx); err != nil {
				n.log.Error("Can't select and sent article", "err", err.Error())
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}
}

func (n *Notifier) SaveArticleFromUser(ctx context.Context, article models.Article) error {
	const op = "services.notifier.save_article_from_user"

	if err := n.articles.Save(ctx, article); err != nil {
		n.log.Error("Can't save article", "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (n *Notifier) SelectPostedArticles(ctx context.Context, limit int64) ([]models.Article, error) {
	const op = "services.notifier.select_posted_articles"

	articles, err := n.articles.LatestPosted(ctx, limit)
	if err != nil {
		if errors.Is(err, storage.ErrNoLatestArticles) {
			n.log.Debug("Can't get latest articles", "err", err.Error())
			return nil, services.ErrNoLatestArticles
		}
		n.log.Error("Can't get latest articles", "err", err.Error())
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return articles, nil
}

func (n *Notifier) selectAndSendArticle(ctx context.Context) error {
	const op = "services.notifier.select_and_send_article"

	article, err := n.articles.NewestNotPosted(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrNoNewArticles) {
			n.log.Debug("Can't get last article", "err", err.Error())
			return services.ErrNoNewArticles
		}
		n.log.Error("Can't get last article", "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := n.articles.MarkPosted(ctx, article.ID); err != nil {
		n.log.Error("Can't mark article as posted", "err", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
