package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"newsWebApp/app/apiService/internal/models"

	chiprometheus "github.com/766b/chi-prometheus"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type AuthService interface {
	SaveUser(ctx context.Context, userName string, email string, password string) (int64, error)
	LoginUser(ctx context.Context, email, password string) (int64, string, string, string, error)
	Parse(ctx context.Context, acToken string) (int64, string, error)
	Refresh(ctx context.Context, refToken string) (int64, string, string, string, error)
	CheckEmail(ctx context.Context, email string) (bool, error)
	CheckUserName(ctx context.Context, userName string) (bool, error)
}

type UserNewsService interface {
	GetArticlesByUid(ctx context.Context, userID int64) ([]models.Article, error)
	SaveArticle(ctx context.Context, userID int64, link string) ([]models.Article, error)
	UpdateArticle(ctx context.Context, userID int64, artID int64, link string) ([]models.Article, error)
	DeleteArticle(ctx context.Context, userID int64, artID int64) ([]models.Article, error)
}

type NewsFetcher interface {
	FetchArticlesOnPage(ctx context.Context, page int64) ([]models.Article, error)
}

func New(auth AuthService,
	news UserNewsService,
	fetcher NewsFetcher,

	refTokTTL time.Duration,
	refreshInterval time.Duration,
	timeout time.Duration,
	slog *slog.Logger,
) (http.Handler, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(chiprometheus.NewMiddleware("apisrv"))
	r.Use(loggerMw(slog))
	r.Use(corsSettings())
	r.Use(refresh(timeout, refTokTTL, auth, slog))

	r.Get("/home", home(timeout, fetcher, slog))
	r.Post("/signup", signup(timeout, auth, slog))
	r.Post("/login", login(timeout, refTokTTL, auth, slog))

	r.Post("/check/email", checkEmail(timeout, auth, slog))
	r.Post("/check/user_name", checkUserName(timeout, auth, slog))

	r.Route("/user_news", func(r chi.Router) {
		r.Use(authenticate(timeout, refTokTTL, auth, slog))
		r.Get("/", userArticles(timeout, news, slog))
		r.Post("/", addArticle(timeout, news, slog))
		r.Put("/", updateArticle(timeout, news, slog))
		r.Delete("/", deleteArticle(timeout, news, slog))
	})

	r.Handle("/metrics", promhttp.Handler())

	return r, nil
}
