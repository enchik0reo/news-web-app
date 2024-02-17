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
	GetArticlesByUid(ctx context.Context, userID int64) ([]models.Article, error)
	SaveArticleFromUser(ctx context.Context, userID int64, link string) ([]models.Article, error)
	UpdateArticleByID(ctx context.Context, userID int64, artID int64, link string) ([]models.Article, error)
	DeleteArticleByID(ctx context.Context, userID int64, artID int64) ([]models.Article, error)
	SelectAndSendArticle(ctx context.Context) (*models.Article, error)
	SelectPostedArticles(ctx context.Context) ([]models.Article, error)
	SelectPostedArticlesWithLimit(ctx context.Context, page int64) ([]models.Article, error)
}

type serverAPI struct {
	newsv1.UnimplementedNewsServer
	newsService NewsService
}

func Register(grpcSrv *grpc.Server, nS NewsService) {
	newsv1.RegisterNewsServer(grpcSrv, &serverAPI{newsService: nS})
}

func (s *serverAPI) GetArticlesByUid(ctx context.Context, req *newsv1.GetArticlesByUidRequest) (*newsv1.GetArticlesByUidResponse, error) {
	articles, err := s.newsService.GetArticlesByUid(ctx, req.GetUserId())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNoOfferedArticles):
			return nil, status.Error(codes.NotFound, "there are no offered articles")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	grpcArticles := make([]*newsv1.Article, len(articles))

	for i, art := range articles {
		grpcArticles[i] = &newsv1.Article{
			ArticleId:  art.ID,
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageUrl:   art.ImageURL,
		}
	}

	return &newsv1.GetArticlesByUidResponse{
		Articles: grpcArticles,
	}, nil
}

func (s *serverAPI) SaveArticle(ctx context.Context, req *newsv1.SaveArticleRequest) (*newsv1.SaveArticleResponse, error) {
	articles, err := s.newsService.SaveArticleFromUser(ctx, req.GetUserId(), req.GetLink())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrArticleSkipped):
			return nil, status.Error(codes.InvalidArgument, "invalid article")
		case errors.Is(err, services.ErrArticleExists):
			return nil, status.Error(codes.AlreadyExists, "article already exists")
		case errors.Is(err, services.ErrNoOfferedArticles):
			return nil, status.Error(codes.NotFound, "there are no offered articles")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	grpcArticles := make([]*newsv1.Article, len(articles))

	for i, art := range articles {
		grpcArticles[i] = &newsv1.Article{
			ArticleId:  art.ID,
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageUrl:   art.ImageURL,
		}
	}

	return &newsv1.SaveArticleResponse{
		Articles: grpcArticles,
	}, nil
}

func (s *serverAPI) UpdateArticle(ctx context.Context, req *newsv1.UpdateArticleRequest) (*newsv1.UpdateArticleResponse, error) {
	articles, err := s.newsService.UpdateArticleByID(ctx, req.GetUserId(), req.GetArticleId(), req.GetLink())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrArticleSkipped):
			return nil, status.Error(codes.InvalidArgument, "invalid article")
		case errors.Is(err, services.ErrArticleExists):
			return nil, status.Error(codes.AlreadyExists, "article already exists")
		case errors.Is(err, services.ErrNoOfferedArticles):
			return nil, status.Error(codes.NotFound, "there are no offered articles")
		case errors.Is(err, services.ErrArticleNotAvailable):
			return nil, status.Error(codes.Unavailable, "there are no changed articles")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	grpcArticles := make([]*newsv1.Article, len(articles))

	for i, art := range articles {
		grpcArticles[i] = &newsv1.Article{
			ArticleId:  art.ID,
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageUrl:   art.ImageURL,
		}
	}

	return &newsv1.UpdateArticleResponse{
		Articles: grpcArticles,
	}, nil
}

func (s *serverAPI) DeleteArticle(ctx context.Context, req *newsv1.DeleteArticleRequest) (*newsv1.DeleteArticleResponse, error) {
	articles, err := s.newsService.DeleteArticleByID(ctx, req.GetUserId(), req.GetArticleId())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNoOfferedArticles):
			return nil, status.Error(codes.NotFound, "there are no offered articles")
		case errors.Is(err, services.ErrArticleNotAvailable):
			return nil, status.Error(codes.Unavailable, "there are no changed articles")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	grpcArticles := make([]*newsv1.Article, len(articles))

	for i, art := range articles {
		grpcArticles[i] = &newsv1.Article{
			ArticleId:  art.ID,
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageUrl:   art.ImageURL,
		}
	}

	return &newsv1.DeleteArticleResponse{
		Articles: grpcArticles,
	}, nil
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
		ArticleId:  art.ID,
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
			ArticleId:  art.ID,
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

func (s *serverAPI) GetArticlesByPage(ctx context.Context, req *newsv1.GetArticlesByPageRequest) (*newsv1.GetArticlesByPageResponse, error) {
	articles, err := s.newsService.SelectPostedArticlesWithLimit(ctx, req.Page)
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
			ArticleId:  art.ID,
			UserName:   art.UserName,
			SourceName: art.SourceName,
			Title:      art.Title,
			Link:       art.Link,
			Excerpt:    art.Excerpt,
			ImageUrl:   art.ImageURL,
			PostedAt:   art.PostedAt.Format(time.DateTime),
		}
	}

	return &newsv1.GetArticlesByPageResponse{
		Articles: grpcArticles,
	}, nil
}
