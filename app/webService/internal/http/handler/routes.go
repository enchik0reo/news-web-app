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
	var ctxKeyArticle = models.ContextKeyArticle("article")

	templatesCache, err := newTemplateCache(templPath)
	if err != nil {
		return nil, fmt.Errorf("can't create new template cache: %w", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(loggerMw(slog))
	r.Use(secureHeaders())
	r.Use(session.Enable)

	if err := fileServer(r, "/static/", templPath); err != nil {
		return nil, fmt.Errorf("can't load static files: %w", err)
	}

	r.Get("/", home(auth, fetcher, templatesCache, session, slog))

	r.Get("/signup", signupForm(auth, templatesCache, session, slog))
	r.Post("/signup", signup(auth, templatesCache, session, slog))
	r.Get("/login", loginForm(auth, templatesCache, session, slog))
	r.Post("/login", login(auth, templatesCache, session, slog))

	r.Route("/logout", func(r chi.Router) {
		//r.Use(requireAuthenticatedUser(auth))
		r.Post("/", logout(session))
	})

	r.Route("/article", func(r chi.Router) {
		//r.Use(requireAuthenticatedUser(auth))
		r.Get("/suggest", suggestArticleForm(auth, templatesCache, session, slog))
		r.Post("/suggest", suggestArticle(auth, news, templatesCache, ctxKeyArticle, session, slog))
		r.Get("/suggested", showArticle(auth, templatesCache, ctxKeyArticle, session, slog))
	})

	r.Get("/refresh", refresh(auth, templatesCache, session, slog))

	return r, nil
}
