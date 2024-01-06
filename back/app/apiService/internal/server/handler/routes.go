package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"newsWebApp/app/apiService/internal/models"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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

func New(auth AuthService,
	news NewsSaver,
	fetcher NewsFetcher,

	refTokTTL time.Duration,
	slog *slog.Logger,
) (http.Handler, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(loggerMw(slog))
	r.Use(corsSettings())
	r.Use(refresh(refTokTTL, auth, slog))

	r.Get("/home", home(auth, fetcher, slog))
	r.Post("/signup", signup(auth, slog))
	r.Post("/login", login(refTokTTL, auth, slog))

	r.Route("/suggest", func(r chi.Router) {
		r.Use(authenticate(refTokTTL, auth, slog))
		r.Post("/", suggestArticle(auth, news, slog))
	})

	return r, nil
}
