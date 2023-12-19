package handler

import (
	"context"
	"log/slog"
	"net/http"
	"newsWebApp/app/webService/internal/models"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type AuthService interface {
	SaveUser(ctx context.Context, userName string, email string, password string) (int64, error)
	LoginUser(ctx context.Context, email, password string) (int64, string, string, string, error)
	Parse(ctx context.Context, acToken string) (int64, string, error)
	Refresh(ctx context.Context, refToken string) (int64, string, string, string, error)
}

type NewsSaver interface {
	SaveArticle(ctx context.Context, userID int64, link string) error
}

type NewsFetcher interface {
	FetchArticles(ctx context.Context) ([]models.Article, error)
}

func New(auth AuthService, news NewsSaver, fetcher NewsFetcher, log *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(loggerMw(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Get("/", index())
	r.Post("/signup", signUp(auth))
	r.Post("/login", login(auth))

	r.Route("/profile", func(r chi.Router) {
		r.Use(authMw(auth))

		r.Get("/", home())
	})

	return r
}
