package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type AuthService interface {
	SaveUser(email string, password string) (int64, error)
	LoginUser(email, password string) (int64, string, string, error)
	Parse(acToken string) (int64, error)
	Refresh(refToken string) (int64, string, string, error)
}

// TODO
type NewsService interface {
	SaveArticle()          // Чтоб сохранять
	GetPostedArticles()    // Чтоб отображать на главной странице опубликованные новости
	AllNotPostedArticles() // Чтоб смотреть не опубликованные новости
	MarkArticlePosted()    // Чтоб пометить новость как опубликованную
}

func New(auth AuthService, log *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(loggerMw(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Get("/", index())
	r.Post("/sign-up", signUp(auth))
	r.Post("/sign-in", signIn(auth))

	r.Route("/profile", func(r chi.Router) {
		r.Use(authMw(auth))

		r.Get("/", home())
	})

	return r
}
