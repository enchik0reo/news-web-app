package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type AuthService interface {
	SaveUser(ctx context.Context, userName string, email string, password string) (int64, error)
	LoginUser(ctx context.Context, email, password string) (int64, string, string, string, error)
	Parse(ctx context.Context, acToken string) (int64, string, error)
	Refresh(ctx context.Context, refToken string) (int64, string, string, string, error)
}

// TODO
type NewsService interface {
	SaveArticle(ctx context.Context)          // Чтоб сохранять
	GetPostedArticles(ctx context.Context)    // Чтоб отображать на главной странице опубликованные новости
	AllNotPostedArticles(ctx context.Context) // Чтоб смотреть не опубликованные новости
	MarkArticlePosted(ctx context.Context)    // Чтоб пометить новость как опубликованную
}

func New(auth AuthService, log *slog.Logger) http.Handler {
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
