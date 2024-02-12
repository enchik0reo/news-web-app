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

	"github.com/gorilla/websocket"
)

func handleConnection(rI time.Duration, upgrader websocket.Upgrader, fetcher NewsFetcher, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Warn("Can't upgrade connection", "err", err.Error())
			return
		}

		go sendNewMsg(rI, wsConn, fetcher, slog)
	}
}

func sendNewMsg(refreshInterval time.Duration, client *websocket.Conn, fetcher NewsFetcher, slog *slog.Logger) {
	ticker := time.NewTicker(refreshInterval / 2)

	defer client.Close()

	for {
		w, err := client.NextWriter(websocket.TextMessage)
		if err != nil {
			ticker.Stop()
			break
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		arts, err := fetcher.FetchArticles(ctx)
		if err != nil {
			ticker.Stop()
			slog.Debug("Can't fetch articles", "err", err.Error())
			break
		}

		resp, err := makeResponse(http.StatusOK, "", "", arts, "")
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}

		w.Write(resp)

		w.Close()

		<-ticker.C
	}
}

func home(fetcher NewsFetcher, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var start = time.Now()
		var statusCode int

		defer func() {
			metrics.ObserveRequest(time.Since(start), statusCode)
		}()

		statusCode = http.StatusOK

		uid := r.Header.Get("uid")
		acToken := r.Header.Get("access_token")

		resp, err := makeResponse(statusCode, uid, acToken, nil, "")
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}

		w.Write(resp)
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

				resp, err := makeResponse(http.StatusBadRequest, "", "", nil, "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if _, err := service.SaveUser(ctx, req.Name, req.Email, req.Password); err != nil {
			switch {
			case errors.Is(err, services.ErrUserExists):
				resp, err := makeResponse(http.StatusNoContent, "", "", nil, "User already exists")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			case errors.Is(err, services.ErrInvalidValue):
				resp, err := makeResponse(http.StatusBadRequest, "", "", nil, "Invalid value")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			default:
				slog.Error("Can't save user", "err", err.Error())

				resp, err := makeResponse(http.StatusInternalServerError, "", "", nil, "Internal error")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			}
		}

		resp, err := makeResponse(http.StatusCreated, "", "", nil, "")
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
		w.Write(resp)
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

				resp, err := makeResponse(http.StatusBadRequest, "", "", nil, "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, _, acsToken, refToken, err := service.LoginUser(ctx, req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserDoesntExists):
				resp, err := makeResponse(http.StatusNoContent, "", "", nil, "Wrong e-mail or password")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			case errors.Is(err, services.ErrInvalidValue):
				resp, err := makeResponse(http.StatusBadRequest, "", "", nil, "Invalid value")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			default:
				slog.Error("Can't login user", "err", err.Error())

				resp, err := makeResponse(http.StatusInternalServerError, "", "", nil, "Internal error")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			}
		}

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

		resp, err := makeResponse(http.StatusAccepted, "", acsToken, nil, "")
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
		w.Write(resp)
	}
}

func userArticles(news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("uid")
		acToken := r.Header.Get("access_token")

		if id == "" {
			slog.Debug("Can't get header id is empty")

			resp, err := makeResponse(http.StatusInternalServerError, id, acToken, nil, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Debug("Can't Atoi", "err", err.Error())

			resp, err := makeResponse(http.StatusInternalServerError, id, acToken, nil, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		arts, err := news.GetArticlesByUid(ctx, int64(uid))
		if err != nil {
			if errors.Is(err, services.ErrNoOfferedArticles) {
				resp, err := makeResponse(http.StatusNoContent, id, acToken, nil, "There are no articles")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			}
			slog.Debug("Can't fetch articles", "err", err.Error())
		}

		if len(arts) == 0 {
			resp, err := makeResponse(http.StatusNoContent, id, acToken, nil, "There are no articles")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		resp, err := makeResponse(http.StatusOK, id, acToken, arts, "")
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
		w.Write(resp)
	}
}

type addRequest struct {
	Link    string `json:"link"`
	Content string `json:"content"`
}

func addArticle(news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(addRequest)

		id := r.Header.Get("uid")
		acToken := r.Header.Get("access_token")

		if id == "" {
			slog.Debug("Can't get header id is empty")

			resp, err := makeResponse(http.StatusInternalServerError, id, acToken, nil, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Debug("Can't Atoi", "err", err.Error())

			resp, err := makeResponse(http.StatusInternalServerError, id, acToken, nil, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from suggest-article request", "err", err.Error())

				resp, err := makeResponse(http.StatusBadRequest, id, acToken, nil, "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
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

				resp, err := makeResponse(http.StatusNoContent, id, acToken, arts, "")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			case errors.Is(err, services.ErrArticleExists):
				slog.Debug("Can't save article", "err", err.Error())

				resp, err := makeResponse(http.StatusPartialContent, id, acToken, arts, "")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			case errors.Is(err, services.ErrNoOfferedArticles):
				resp, err := makeResponse(http.StatusResetContent, id, acToken, arts, "")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			default:
				slog.Error("Can't save article", "err", err.Error())

				resp, err := makeResponse(http.StatusInternalServerError, id, acToken, arts, "")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			}
		}

		if len(arts) == 0 {
			resp, err := makeResponse(http.StatusResetContent, id, acToken, nil, "There are no articles")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		resp, err := makeResponse(http.StatusCreated, id, acToken, arts, "")
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
		w.Write(resp)
	}
}

type updateRequest struct {
	ArticleID int    `json:"article_id"`
	Link      string `json:"link"`
}

func updateArticle(news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(updateRequest)

		id := r.Header.Get("uid")
		acToken := r.Header.Get("access_token")

		if id == "" {
			slog.Debug("Can't get header id is empty")

			resp, err := makeResponse(http.StatusInternalServerError, id, acToken, nil, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Debug("Can't Atoi", "err", err.Error())

			resp, err := makeResponse(http.StatusInternalServerError, id, acToken, nil, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from suggest-article request", "err", err.Error())

				resp, err := makeResponse(http.StatusBadRequest, id, acToken, nil, "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
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

				resp, err := makeResponse(http.StatusNoContent, id, acToken, arts, "")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			case errors.Is(err, services.ErrArticleExists):
				slog.Debug("Can't update article", "err", err.Error())

				resp, err := makeResponse(http.StatusPartialContent, id, acToken, arts, "")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			case errors.Is(err, services.ErrNoOfferedArticles):
				resp, err := makeResponse(http.StatusResetContent, id, acToken, arts, "")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			case errors.Is(err, services.ErrArticleNotAvailable):
				resp, err := makeResponse(http.StatusForbidden, id, acToken, arts, "")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			default:
				slog.Error("Can't update article", "err", err.Error())

				resp, err := makeResponse(http.StatusInternalServerError, id, acToken, arts, "")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			}
		}

		if len(arts) == 0 {
			resp, err := makeResponse(http.StatusResetContent, id, acToken, nil, "There are no articles")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		resp, err := makeResponse(http.StatusAccepted, id, acToken, arts, "")
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
		w.Write(resp)
	}
}

func deleteArticle(news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("uid")
		acToken := r.Header.Get("access_token")

		if id == "" {
			slog.Debug("Can't get header id is empty")

			resp, err := makeResponse(http.StatusInternalServerError, id, acToken, nil, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Debug("Can't convert to int", "err", err.Error())

			resp, err := makeResponse(http.StatusInternalServerError, id, acToken, nil, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		articleIdHeader := r.Header.Get("article_id")

		articleId, err := strconv.Atoi(articleIdHeader)
		if err != nil {
			slog.Debug("Can't convert to int", "err", err.Error())

			resp, err := makeResponse(http.StatusInternalServerError, id, acToken, nil, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		arts, err := news.DeleteArticle(ctx, int64(uid), int64(articleId))
		if err != nil {
			switch {
			case errors.Is(err, services.ErrNoOfferedArticles):
				resp, err := makeResponse(http.StatusNoContent, id, acToken, arts, "")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			case errors.Is(err, services.ErrArticleNotAvailable):
				resp, err := makeResponse(http.StatusAlreadyReported, id, acToken, arts, "")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			default:
				slog.Error("Can't delete article", "err", err.Error())

				resp, err := makeResponse(http.StatusInternalServerError, id, acToken, nil, "Internal error")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				w.Write(resp)
				return
			}
		}

		if len(arts) == 0 {
			resp, err := makeResponse(http.StatusResetContent, id, acToken, nil, "There are no articles")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			w.Write(resp)
			return
		}

		resp, err := makeResponse(http.StatusOK, id, acToken, arts, "")
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
		w.Write(resp)
	}
}
