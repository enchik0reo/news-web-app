package newsgrpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"newsWebApp/app/apiService/internal/models"
	"newsWebApp/app/apiService/internal/services"
	newsv1 "newsWebApp/protos/gen/go/news"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type articlesInformer interface {
	String() string
	GetArticles() []*newsv1.Article
}

type articleInformer interface {
	String() string
	GetArticl() *newsv1.Article
}

type Client struct {
	api newsv1.NewsClient
}

func New(ctx context.Context,
	log *slog.Logger,
	host string,
	port int,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	const op = "services.newsgrpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.DialContext(ctx,
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(interceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: newsv1.NewNewsClient(cc),
	}, nil
}

func (c *Client) GetArticlesByUid(ctx context.Context, userID int64) ([]models.Article, error) {
	const op = "services.newsgrpc.GetArticlesById"

	resp, err := c.api.GetArticlesByUid(ctx, &newsv1.GetArticlesByUidRequest{UserId: userID})
	if err != nil {
		switch {
		case errors.Is(err, status.Error(codes.NotFound, "there are no offered articles")):
			return nil, services.ErrNoOfferedArticles
		default:
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	articles := make([]models.Article, len(resp.Articles))

	for i, art := range resp.Articles {
		articles[i] = models.Article{
			ArticleID:  art.ArticleId,
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageURL:   art.ImageUrl,
		}
	}

	return articles, nil
}

func (c *Client) SaveArticle(ctx context.Context, userID int64, link string) ([]models.Article, error) {
	const op = "services.newsgrpc.SaveArticle"

	resp, err := c.api.SaveArticle(ctx, &newsv1.SaveArticleRequest{UserId: userID, Link: link})
	if err != nil {
		switch {
		case errors.Is(err, status.Error(codes.InvalidArgument, "invalid article")):
			return nil, services.ErrArticleSkipped
		case errors.Is(err, status.Error(codes.AlreadyExists, "article already exists")):
			return nil, services.ErrArticleExists
		case errors.Is(err, status.Error(codes.NotFound, "there are no offered articles")):
			return nil, services.ErrNoOfferedArticles
		case errors.Is(err, status.Error(codes.Unknown, "url is invalid")):
			return nil, services.ErrInvalidUrl
		default:
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	articles := make([]models.Article, len(resp.Articles))

	for i, art := range resp.Articles {
		articles[i] = models.Article{
			ArticleID:  art.ArticleId,
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageURL:   art.ImageUrl,
		}
	}

	return articles, nil
}

func (c *Client) UpdateArticle(ctx context.Context, userID int64, artID int64, link string) ([]models.Article, error) {
	const op = "services.newsgrpc.UpdateArticle"

	resp, err := c.api.UpdateArticle(ctx, &newsv1.UpdateArticleRequest{UserId: userID, ArticleId: artID, Link: link})
	if err != nil {
		switch {
		case errors.Is(err, status.Error(codes.InvalidArgument, "invalid article")):
			return nil, services.ErrArticleSkipped
		case errors.Is(err, status.Error(codes.AlreadyExists, "article already exists")):
			return nil, services.ErrArticleExists
		case errors.Is(err, status.Error(codes.NotFound, "there are no offered articles")):
			return nil, services.ErrNoOfferedArticles
		case errors.Is(err, status.Error(codes.Unavailable, "there are no changed articles")):
			return nil, services.ErrArticleNotAvailable
		case errors.Is(err, status.Error(codes.Unknown, "url is invalid")):
			return nil, services.ErrInvalidUrl
		default:
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	articles := make([]models.Article, len(resp.Articles))

	for i, art := range resp.Articles {
		articles[i] = models.Article{
			ArticleID:  art.ArticleId,
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageURL:   art.ImageUrl,
		}
	}

	return articles, nil
}

func (c *Client) DeleteArticle(ctx context.Context, userID int64, artID int64) ([]models.Article, error) {
	const op = "services.newsgrpc.DeleteArticle"

	resp, err := c.api.DeleteArticle(ctx, &newsv1.DeleteArticleRequest{UserId: userID, ArticleId: artID})
	if err != nil {
		switch {
		case errors.Is(err, status.Error(codes.NotFound, "there are no offered articles")):
			return nil, services.ErrNoOfferedArticles
		case errors.Is(err, status.Error(codes.Unavailable, "there are no changed articles")):
			return nil, services.ErrArticleNotAvailable
		default:
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	articles := make([]models.Article, len(resp.Articles))

	for i, art := range resp.Articles {
		articles[i] = models.Article{
			ArticleID:  art.ArticleId,
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageURL:   art.ImageUrl,
		}
	}

	return articles, nil
}

func (c *Client) GetNewestArticle(ctx context.Context) (*models.Article, error) {
	const op = "services.newsgrpc.GetNewestArticle"

	resp, err := c.api.GetNewestArticle(ctx, &newsv1.GetNewestArticleRequest{})
	if err != nil {
		if errors.Is(err, status.Error(codes.NotFound, "there is no new article")) {
			return nil, fmt.Errorf("%s: %w", op, services.ErrNoNewArticle)
		} else {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	art := resp.Articl

	postedAt, err := time.Parse(time.DateTime, art.PostedAt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	article := models.Article{
		ArticleID:  art.ArticleId,
		UserName:   art.UserName,
		SourceName: art.SourceName,
		Title:      art.Title,
		Link:       art.Link,
		Excerpt:    art.Excerpt,
		ImageURL:   art.ImageUrl,
		PostedAt:   postedAt,
	}

	return &article, nil
}

func (c *Client) GetArticles(ctx context.Context) ([]models.Article, error) {
	const op = "services.newsgrpc.GetArticles"

	resp, err := c.api.GetArticles(ctx, &newsv1.GetArticlesRequest{})
	if err != nil {
		if errors.Is(err, status.Error(codes.NotFound, "there are no published articles")) {
			return nil, fmt.Errorf("%s: %w", op, services.ErrNoPublishedArticles)
		} else {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	articles := make([]models.Article, len(resp.Articles))

	for i, art := range resp.Articles {
		postedAt, err := time.Parse(time.DateTime, art.PostedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		articles[i] = models.Article{
			ArticleID:  art.ArticleId,
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageURL:   art.ImageUrl,
			PostedAt:   postedAt,
		}
	}

	return articles, nil
}

func (c *Client) GetArticlesByPage(ctx context.Context, page int64) ([]models.Article, error) {
	const op = "services.newsgrpc.GetArticles"

	resp, err := c.api.GetArticlesByPage(ctx, &newsv1.GetArticlesByPageRequest{Page: page})
	if err != nil {
		if errors.Is(err, status.Error(codes.NotFound, "there are no published articles")) {
			return nil, fmt.Errorf("%s: %w", op, services.ErrNoPublishedArticles)
		} else {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	articles := make([]models.Article, len(resp.Articles))

	for i, art := range resp.Articles {
		postedAt, err := time.Parse(time.DateTime, art.PostedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		articles[i] = models.Article{
			ArticleID:  art.ArticleId,
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageURL:   art.ImageUrl,
			PostedAt:   postedAt,
		}
	}

	return articles, nil
}

func interceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		grpcFields := grpclog.Fields(fields)
		iterator := grpcFields.Iterator()

		contentLen := 0
		contentCount := 0

	Loop:
		for iterator.Next() {
			k, v := iterator.At()
			if k == "grpc.response.content" {
				switch resp := v.(type) {
				case articlesInformer:
					contentLen = len(resp.String())
					contentCount = len(resp.GetArticles())
					grpcFields.Delete("grpc.response.content")
					break Loop
				case articleInformer:
					contentLen = len(resp.String())
					contentCount = 1
					grpcFields.Delete("grpc.response.content")
					break Loop
				default:
					break Loop
				}
			}
		}

		if contentCount > 0 {
			grpcFields = grpcFields.AppendUnique(grpclog.Fields{"grpc.response.content", map[string]interface{}{
				"Articles count": fmt.Sprintf("%d", contentCount),
				"Total lenght":   fmt.Sprintf("%d bytes", contentLen),
			}})
		}

		l.Log(ctx, slog.Level(level), msg, grpcFields...)
	})
}
