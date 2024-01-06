package handler

import (
	"context"
	"errors"
	"time"

	"newsWebApp/app/newsService/internal/models"
	"newsWebApp/app/newsService/internal/services"

	newsv1 "newsWebApp/protos/gen/go/news"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NewsService interface {
	SaveArticleFromUser(ctx context.Context, userID int64, link string) error
	SelectAndSendArticle(ctx context.Context) (*models.Article, error)
	SelectPostedArticles(ctx context.Context) ([]models.Article, error)
}

type serverAPI struct {
	newsv1.UnimplementedNewsServer
	newsService NewsService
}

func Register(grpcSrv *grpc.Server, nS NewsService) {
	newsv1.RegisterNewsServer(grpcSrv, &serverAPI{newsService: nS})
}

func (s *serverAPI) SaveArticle(ctx context.Context, req *newsv1.SaveArticleRequest) (*newsv1.SaveArticleResponse, error) {
	if err := s.newsService.SaveArticleFromUser(ctx, req.GetUserId(), req.GetLink()); err != nil {
		switch {
		case errors.Is(err, services.ErrArticleSkipped):
			return nil, status.Error(codes.InvalidArgument, "invalid article")
		case errors.Is(err, services.ErrArticleExists):
			return nil, status.Error(codes.AlreadyExists, "article already exists")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &newsv1.SaveArticleResponse{}, nil
}

func (s *serverAPI) GetNewestArticle(ctx context.Context, req *newsv1.GetNewestArticleRequest) (*newsv1.GetNewestArticleResponse, error) {
	art, err := s.newsService.SelectAndSendArticle(ctx)
	if err != nil {
		if errors.Is(err, services.ErrNoNewArticle) {
			return nil, status.Error(codes.NotFound, "there is no new article")
		} else {
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	grpcArticle := &newsv1.Article{
		UserName:   art.UserName,
		SourceName: art.SourceName,
		Title:      art.Title,
		Link:       art.Link,
		Excerpt:    art.Excerpt,
		ImageUrl:   art.ImageURL,
		PostedAt:   art.PostedAt.Format(time.DateTime),
	}

	return &newsv1.GetNewestArticleResponse{Articl: grpcArticle}, nil
}

func (s *serverAPI) GetArticles(ctx context.Context, req *newsv1.GetArticlesRequest) (*newsv1.GetArticlesResponse, error) {
	articles, err := s.newsService.SelectPostedArticles(ctx)
	if err != nil {
		if errors.Is(err, services.ErrNoPublishedArticles) {
			return nil, status.Error(codes.NotFound, "there are no published articles")
		} else {
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	grpcArticles := make([]*newsv1.Article, len(articles))

	for i, art := range articles {
		grpcArticles[i] = &newsv1.Article{
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageUrl:   art.ImageURL,
			PostedAt:   art.PostedAt.Format(time.DateTime),
		}
	}

	return &newsv1.GetArticlesResponse{
		Articles: grpcArticles,
	}, nil
}
