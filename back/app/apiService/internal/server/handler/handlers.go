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

	"newsWebApp/app/apiService/internal/services"
)

func home(service AuthService, fetcher NewsFetcher, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		arts, err := fetcher.FetchArticles(ctx)
		if err != nil {
			slog.Debug("Can't fetch articles", "err", err.Error())
		}

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

type suggestRequest struct {
	Link    string `json:"link"`
	Content string `json:"content"`
}

func suggestArticle(service AuthService, news NewsSaver, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(suggestRequest)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from suggest-article request", "err", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

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

		if err := news.SaveArticle(ctx, int64(uid), req.Link); err != nil {
			switch {
			case errors.Is(err, services.ErrArticleSkipped):
				slog.Debug("Can't save article", "err", err.Error())
				w.WriteHeader(http.StatusNoContent)
				return
			case errors.Is(err, services.ErrArticleExists):
				slog.Debug("Can't save article", "err", err.Error())
				w.WriteHeader(http.StatusPartialContent)
				return
			default:
				slog.Error("Can't save article", "err", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
	}
}
