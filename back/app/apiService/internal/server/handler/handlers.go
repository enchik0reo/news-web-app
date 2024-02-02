package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"newsWebApp/app/apiService/internal/metrics"
	"newsWebApp/app/apiService/internal/services"
)

func home(fetcher NewsFetcher, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		statusCode := 0
		defer func() {
			metrics.ObserveRequest(time.Since(start), statusCode)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		arts, err := fetcher.FetchArticles(ctx)
		if err != nil {
			slog.Debug("Can't fetch articles", "err", err.Error())
		}

		if len(arts) == 0 {
			statusCode = http.StatusNoContent
			w.WriteHeader(statusCode)
			return
		}

		statusCode = http.StatusOK
		w.WriteHeader(statusCode)

		if err := json.NewEncoder(w).Encode(arts); err != nil {
			slog.Error("Can't encode articles", "err", err.Error())
		}
	}
}

type signUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func signup(service AuthService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := signUpRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from sign-up request", "err", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if _, err := service.SaveUser(ctx, req.Name, req.Email, req.Password); err != nil {
			switch {
			case errors.Is(err, services.ErrUserExists):
				w.WriteHeader(http.StatusNoContent)
				return
			case errors.Is(err, services.ErrInvalidValue):
				w.WriteHeader(http.StatusBadRequest)
				return
			default:
				slog.Error("Can't save user", "err", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func login(refTokTTL time.Duration, service AuthService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(loginRequest)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from login request", "err", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, _, acsToken, refToken, err := service.LoginUser(ctx, req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserDoesntExists):
				w.WriteHeader(http.StatusNoContent)
				return
			case errors.Is(err, services.ErrInvalidValue):
				w.WriteHeader(http.StatusBadRequest)
				return
			default:
				slog.Error("Can't login user", "err", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("access_token", acsToken)

		ck := http.Cookie{
			Name:     "refresh_token",
			Domain:   r.URL.Host,
			Path:     "/",
			Value:    refToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(refTokTTL),
		}

		http.SetCookie(w, &ck)

		w.WriteHeader(http.StatusAccepted)
	}
}

func userArticles(news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := w.Header().Get("id")
		if id == "" {
			slog.Error("Can't get header id is empty")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Debug("Can't Atoi", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		arts, err := news.GetArticlesByUid(ctx, int64(uid))
		if err != nil {
			if errors.Is(err, services.ErrNoOfferedArticles) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			slog.Debug("Can't fetch articles", "err", err.Error())
		}

		if len(arts) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(arts); err != nil {
			slog.Error("Can't encode articles", "err", err.Error())
		}
	}
}

type addRequest struct {
	Link    string `json:"link"`
	Content string `json:"content"`
}

func addArticle(news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(addRequest)

		id := w.Header().Get("id")
		if id == "" {
			slog.Error("Can't get header id is empty")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Debug("Can't Atoi", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from suggest-article request", "err", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		arts, err := news.SaveArticle(ctx, int64(uid), req.Link)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrArticleSkipped):
				slog.Debug("Can't save article", "err", err.Error())
				w.WriteHeader(http.StatusNoContent)
				return
			case errors.Is(err, services.ErrArticleExists):
				slog.Debug("Can't save article", "err", err.Error())
				w.WriteHeader(http.StatusPartialContent)
				return
			case errors.Is(err, services.ErrNoOfferedArticles):
				w.WriteHeader(http.StatusResetContent)
				return
			default:
				slog.Error("Can't save article", "err", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if len(arts) == 0 {
			w.WriteHeader(http.StatusResetContent)
			return
		}

		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(arts); err != nil {
			slog.Error("Can't encode articles", "err", err.Error())
		}
	}
}

type updateRequest struct {
	ArticleID int    `json:"article_id"`
	Link      string `json:"link"`
}

func updateArticle(news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(updateRequest)

		id := w.Header().Get("id")
		if id == "" {
			slog.Error("Can't get header id is empty")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Debug("Can't Atoi", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from suggest-article request", "err", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		arts, err := news.UpdateArticle(ctx, int64(uid), int64(req.ArticleID), req.Link)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrArticleSkipped):
				slog.Debug("Can't update article", "err", err.Error())
				w.WriteHeader(http.StatusNoContent)
				return
			case errors.Is(err, services.ErrArticleExists):
				slog.Debug("Can't update article", "err", err.Error())
				w.WriteHeader(http.StatusPartialContent)
				return
			case errors.Is(err, services.ErrNoOfferedArticles):
				w.WriteHeader(http.StatusResetContent)
				return
			default:
				slog.Error("Can't update article", "err", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if len(arts) == 0 {
			w.WriteHeader(http.StatusResetContent)
			return
		}

		w.WriteHeader(http.StatusAccepted)

		if err := json.NewEncoder(w).Encode(arts); err != nil {
			slog.Error("Can't encode articles", "err", err.Error())
		}
	}
}

func deleteArticle(news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := w.Header().Get("id")
		if id == "" {
			slog.Error("Can't get header id is empty")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Error("Can't Atoi", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		articleIdHeader := r.Header.Get("article_id")

		articleId, err := strconv.Atoi(articleIdHeader)
		if err != nil {
			slog.Error("Can't Atoi", "err", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		arts, err := news.DeleteArticle(ctx, int64(uid), int64(articleId))
		if err != nil {
			switch {
			case errors.Is(err, services.ErrNoOfferedArticles):
				w.WriteHeader(http.StatusNoContent)
				return
			default:
				slog.Error("Can't delete article", "err", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if len(arts) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(arts); err != nil {
			slog.Error("Can't encode articles", "err", err.Error())
		}
	}
}
