package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"newsWebApp/app/webService/internal/models"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/golangcollege/sessions"
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
	templPath string,
	session *sessions.Session,
	slog *slog.Logger,
) (http.Handler, error) {
	var ctxKeyUser = models.ContextKey("user")
	var ctxKeyArticle = models.ContextKeyArticle("article")

	templatesCache, err := newTemplateCache(templPath)
	if err != nil {
		return nil, fmt.Errorf("can't create new template cache: %w", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(loggerMw(slog))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(session.Enable)
	r.Use(authenticate(ctxKeyUser, auth))

	if err := fileServer(r, "/static/", templPath); err != nil {
		return nil, fmt.Errorf("can't load static files: %w", err)
	}

	r.Get("/", home(fetcher, templatesCache, ctxKeyUser, session))

	r.Get("/signup", signupForm(templatesCache, ctxKeyUser, session))
	r.Post("/signup", signup(auth, templatesCache, ctxKeyUser, session))
	r.Get("/login", loginForm(templatesCache, ctxKeyUser, session))
	r.Post("/login", login(auth, templatesCache, ctxKeyUser, session))

	r.Route("/logout", func(r chi.Router) {
		r.Use(requireAuthenticatedUser(ctxKeyUser, auth))
		r.Post("/", logout(ctxKeyUser, session))
	})

	r.Route("/article", func(r chi.Router) {
		r.Use(requireAuthenticatedUser(ctxKeyUser, auth))
		r.Get("/suggest", suggestArticleForm(templatesCache, ctxKeyUser, session))
		r.Post("/suggest", suggestArticle(news, templatesCache, ctxKeyUser, ctxKeyArticle, session))
		r.Get("/suggested", showArticle(templatesCache, ctxKeyUser, ctxKeyArticle, session))
	})

	r.Route("/check", func(r chi.Router) {
		r.Get("/", whoami(ctxKeyUser))
	})

	r.Get("/ping", ping())

	return r, nil
}
