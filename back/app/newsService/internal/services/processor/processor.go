package processor

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
	NewestNotPosted(ctx context.Context) (*models.Article, error)
	LatestPosted(ctx context.Context) ([]models.Article, error)
	LatestPostedWithLimit(ctx context.Context, limit int64) ([]models.Article, error)
	MarkPosted(ctx context.Context, id int64) (time.Time, error)
	ArticlesByUid(ctx context.Context, userID int64) ([]models.Article, error)
}

type UserArticleStorage interface {
	SaveArticleFromUser(ctx context.Context, userID int64, link string) error
	UpdateArticleByID(ctx context.Context, userID int64, artID int64, link string) error
	DeleteArticleByID(ctx context.Context, userID int64, artID int64) error
}

type Processor struct {
	articles ArticleStorage
	uArt     UserArticleStorage

	pageLimit int
	log       *slog.Logger
}

func New(
	articles ArticleStorage,
	uArt UserArticleStorage,
	articlesLimit int,
	log *slog.Logger,
) *Processor {
	return &Processor{
		articles:  articles,
		uArt:      uArt,
		pageLimit: articlesLimit,
		log:       log,
	}
}

func (p *Processor) GetArticlesByUid(ctx context.Context, userID int64) ([]models.Article, error) {
	const op = "services.processor.get_articles_by_id"

	articles, err := p.articles.ArticlesByUid(ctx, userID)
	if err != nil || len(articles) == 0 {
		switch {
		case len(articles) == 0:
			p.log.Debug("There are no offered articles")
			return nil, services.ErrNoOfferedArticles
		default:
			p.log.Error("Can't get articles by id", "err", err.Error())
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return articles, nil
}

func (p *Processor) SaveArticleFromUser(ctx context.Context, userID int64, link string) ([]models.Article, error) {
	const op = "services.processor.save_article_from_user"

	if err := p.uArt.SaveArticleFromUser(ctx, userID, link); err != nil {
		switch {
		case errors.Is(err, services.ErrArticleSkipped):
			p.log.Debug("Can't save article from user", "err", err.Error())
			return nil, services.ErrArticleSkipped
		case errors.Is(err, services.ErrArticleExists):
			p.log.Debug("Can't save article from user", "err", err.Error())
			return nil, services.ErrArticleExists
		default:
			p.log.Error("Can't save article from user", "err", err.Error())
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	articles, err := p.articles.ArticlesByUid(ctx, userID)
	if err != nil || len(articles) == 0 {
		switch {
		case len(articles) == 0:
			p.log.Debug("There are no offered articles")
			return nil, services.ErrNoOfferedArticles
		default:
			p.log.Error("Can't get articles by id", "err", err.Error())
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return articles, nil
}

func (p *Processor) UpdateArticleByID(ctx context.Context, userID int64, artID int64, link string) ([]models.Article, error) {
	const op = "services.processor.update_article_by_id"

	if err := p.uArt.UpdateArticleByID(ctx, userID, artID, link); err != nil {
		switch {
		case errors.Is(err, services.ErrArticleSkipped):
			p.log.Debug("Can't update article from user", "err", err.Error())
			return nil, services.ErrArticleSkipped
		case errors.Is(err, services.ErrArticleExists):
			p.log.Debug("Can't update article from user", "err", err.Error())
			return nil, services.ErrArticleExists
		case errors.Is(err, services.ErrArticleNotAvailable):
			p.log.Debug("Can't update article from user", "err", err.Error())
			return nil, services.ErrArticleNotAvailable
		default:
			p.log.Error("Can't update article from user", "err", err.Error())
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	articles, err := p.articles.ArticlesByUid(ctx, userID)
	if err != nil || len(articles) == 0 {
		switch {
		case len(articles) == 0:
			p.log.Debug("There are no offered articles")
			return nil, services.ErrNoOfferedArticles
		default:
			p.log.Error("Can't get articles by id", "err", err.Error())
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return articles, nil
}

func (p *Processor) DeleteArticleByID(ctx context.Context, userID int64, artID int64) ([]models.Article, error) {
	const op = "services.processor.delete_article_by_id"

	if err := p.uArt.DeleteArticleByID(ctx, userID, artID); err != nil {
		if errors.Is(err, services.ErrArticleNotAvailable) {
			p.log.Debug("Can't delete article from user", "err", err.Error())
			return nil, services.ErrArticleNotAvailable
		}
		p.log.Error("Can't delete articles by id", "err", err.Error())
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	articles, err := p.articles.ArticlesByUid(ctx, userID)
	if err != nil || len(articles) == 0 {
		switch {
		case len(articles) == 0:
			p.log.Debug("There are no offered articles")
			return nil, services.ErrNoOfferedArticles
		default:
			p.log.Error("Can't get articles by id", "err", err.Error())
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return articles, nil
}

func (p *Processor) SelectPostedArticles(ctx context.Context) ([]models.Article, error) {
	const op = "services.processor.select_posted_articles"

	articles, err := p.articles.LatestPosted(ctx)
	if err != nil || len(articles) == 0 {
		switch {
		case len(articles) == 0:
			p.log.Debug("There are no latest posted articles")
			return nil, services.ErrNoPublishedArticles
		case errors.Is(err, storage.ErrNoSources):
			p.log.Debug("Can't get latest posted articles", "err", err.Error())
			return nil, services.ErrNoPublishedArticles
		default:
			p.log.Error("Can't get latest posted articles", "err", err.Error())
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return articles, nil
}

func (p *Processor) SelectPostedArticlesWithLimit(ctx context.Context, page int64) ([]models.Article, error) {
	const op = "services.processor.select_posted_articles"

	limit := page * int64(p.pageLimit)

	articles, err := p.articles.LatestPostedWithLimit(ctx, limit)
	if err != nil || len(articles) == 0 {
		switch {
		case len(articles) == 0:
			p.log.Debug("There are no latest posted articles")
			return nil, services.ErrNoPublishedArticles
		case errors.Is(err, storage.ErrNoSources):
			p.log.Debug("Can't get latest posted articles", "err", err.Error())
			return nil, services.ErrNoPublishedArticles
		default:
			p.log.Error("Can't get latest posted articles", "err", err.Error())
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return articles, nil
}

func (p *Processor) SelectAndSendArticle(ctx context.Context) (*models.Article, error) {
	const op = "services.processor.select_and_send_article"

	article, err := p.articles.NewestNotPosted(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrNoNewArticles) {
			p.log.Debug("Can't get last article", "err", err.Error())
			return nil, services.ErrNoNewArticle
		}
		p.log.Error("Can't get last article", "err", err.Error())
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	postedAt, err := p.articles.MarkPosted(ctx, article.ID)
	if err != nil {
		p.log.Error("Can't mark article as posted", "err", err.Error())
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	article.PostedAt = postedAt

	return article, nil
}
