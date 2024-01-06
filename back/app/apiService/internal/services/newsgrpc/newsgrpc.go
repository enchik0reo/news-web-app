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

type Client struct {
	api newsv1.NewsClient
}

func New(ctx context.Context,
	log *slog.Logger,
	addr int,
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
		fmt.Sprintf(":%d", addr),
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

func (c *Client) SaveArticle(ctx context.Context, userID int64, link string) error {
	const op = "services.newsgrpc.SaveArticle"

	if _, err := c.api.SaveArticle(ctx, &newsv1.SaveArticleRequest{UserId: userID, Link: link}); err != nil {
		switch {
		case errors.Is(err, status.Error(codes.InvalidArgument, "invalid article")):
			return services.ErrArticleSkipped
		case errors.Is(err, status.Error(codes.AlreadyExists, "article already exists")):
			return services.ErrArticleExists
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
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

func interceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}
