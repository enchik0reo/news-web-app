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

func handleConnection(timeout, rI time.Duration, upgrader websocket.Upgrader, fetcher NewsFetcher, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Warn("Can't upgrade connection", "err", err.Error())
			return
		}

		w.Header().Add("Content-Type", "application/json")

		go sendNewMsg(timeout, rI, wsConn, fetcher, slog)
	}
}

func sendNewMsg(timeout, refreshInterval time.Duration, client *websocket.Conn, fetcher NewsFetcher, slog *slog.Logger) {
	ticker := time.NewTicker(refreshInterval / 2)

	defer client.Close()

	for {
		w, err := client.NextWriter(websocket.TextMessage)
		if err != nil {
			ticker.Stop()
			break
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		arts, err := fetcher.FetchArticles(ctx)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrNoPublishedArticles):
				err = socketResponse(w, http.StatusOK, arts)
				if err != nil {
					slog.Debug("Can't make socket response", "err", err.Error())
				}
				<-ticker.C
				continue
			case errors.Is(err, context.DeadlineExceeded):
				err = socketResponse(w, http.StatusOK, arts)
				if err != nil {
					slog.Debug("Can't make socket response", "err", err.Error())
				}
				<-ticker.C
				continue
			default:
				slog.Error("Can't fetch articles", "err", err.Error())
				<-ticker.C
				continue
			}
		}

		err = socketResponse(w, http.StatusOK, arts)
		if err != nil {
			slog.Debug("Can't make socket response", "err", err.Error())
		}

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

		err := responseJSON(w, statusCode, uid, acToken, nil)
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}

type signUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func signup(timeout time.Duration, service AuthService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := signUpRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from sign-up request", "err", err.Error())

				err := responseJSONError(w, http.StatusBadRequest, "", "", "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if _, err := service.SaveUser(ctx, req.Name, req.Email, req.Password); err != nil {
			switch {
			case errors.Is(err, services.ErrUserExists):
				err = responseJSONError(w, http.StatusNoContent, "", "", "User already exists")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			case errors.Is(err, services.ErrInvalidValue):
				err = responseJSONError(w, http.StatusBadRequest, "", "", "Invalid value")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			default:
				slog.Error("Can't save user", "err", err.Error())

				err = responseJSONError(w, http.StatusInternalServerError, "", "", "Internal error")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		err := responseJSON(w, http.StatusCreated, "", "", nil)
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func login(timeout time.Duration, refTokTTL time.Duration, service AuthService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(loginRequest)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from login request", "err", err.Error())

				err = responseJSONError(w, http.StatusBadRequest, "", "", "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		_, _, acsToken, refToken, err := service.LoginUser(ctx, req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserDoesntExists):
				err = responseJSONError(w, http.StatusNoContent, "", "", "Wrong e-mail or password")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			case errors.Is(err, services.ErrInvalidValue):
				err = responseJSONError(w, http.StatusBadRequest, "", "", "Invalid value")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			default:
				slog.Error("Can't login user", "err", err.Error())

				err = responseJSONError(w, http.StatusInternalServerError, "", "", "Internal error")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
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

		err = responseJSON(w, http.StatusAccepted, "", acsToken, nil)
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}

func userArticles(timeout time.Duration, news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("uid")
		acToken := r.Header.Get("access_token")

		if id == "" {
			slog.Debug("Can't get header id is empty")

			err := responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Debug("Can't Atoi", "err", err.Error())

			err = responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		arts, err := news.GetArticlesByUid(ctx, int64(uid))
		if err != nil {
			if errors.Is(err, services.ErrNoOfferedArticles) {
				err = responseJSONError(w, http.StatusNoContent, id, acToken, "There are no articles")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
			slog.Debug("Can't fetch articles", "err", err.Error())
		}

		if len(arts) == 0 {
			err = responseJSONError(w, http.StatusNoContent, id, acToken, "There are no articles")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		err = responseJSON(w, http.StatusOK, id, acToken, arts)
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}

type addRequest struct {
	Link    string `json:"link"`
	Content string `json:"content"`
}

func addArticle(timeout time.Duration, news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(addRequest)

		id := r.Header.Get("uid")
		acToken := r.Header.Get("access_token")

		if id == "" {
			slog.Debug("Can't get header id is empty")

			err := responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Debug("Can't Atoi", "err", err.Error())

			err = responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from suggest-article request", "err", err.Error())

				err = responseJSONError(w, http.StatusBadRequest, id, acToken, "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		arts, err := news.SaveArticle(ctx, int64(uid), req.Link)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrArticleSkipped):
				slog.Debug("Can't save article", "err", err.Error())

				err = responseJSON(w, http.StatusNoContent, id, acToken, arts)
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			case errors.Is(err, services.ErrArticleExists):
				slog.Debug("Can't save article", "err", err.Error())

				err = responseJSON(w, http.StatusPartialContent, id, acToken, arts)
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			case errors.Is(err, services.ErrNoOfferedArticles):
				err = responseJSON(w, http.StatusResetContent, id, acToken, arts)
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			default:
				slog.Error("Can't save article", "err", err.Error())

				err = responseJSON(w, http.StatusInternalServerError, id, acToken, arts)
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		if len(arts) == 0 {
			err = responseJSONError(w, http.StatusResetContent, id, acToken, "There are no articles")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		err = responseJSON(w, http.StatusCreated, id, acToken, arts)
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}

type updateRequest struct {
	ArticleID int    `json:"article_id"`
	Link      string `json:"link"`
}

func updateArticle(timeout time.Duration, news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(updateRequest)

		id := r.Header.Get("uid")
		acToken := r.Header.Get("access_token")

		if id == "" {
			slog.Debug("Can't get header id is empty")

			err := responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Debug("Can't Atoi", "err", err.Error())

			err = responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from update-article request", "err", err.Error())

				err = responseJSONError(w, http.StatusBadRequest, id, acToken, "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		arts, err := news.UpdateArticle(ctx, int64(uid), int64(req.ArticleID), req.Link)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrArticleSkipped):
				slog.Debug("Can't update article", "err", err.Error())

				err = responseJSON(w, http.StatusNoContent, id, acToken, arts)
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			case errors.Is(err, services.ErrArticleExists):
				slog.Debug("Can't update article", "err", err.Error())

				err = responseJSON(w, http.StatusPartialContent, id, acToken, arts)
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			case errors.Is(err, services.ErrNoOfferedArticles):
				err = responseJSON(w, http.StatusResetContent, id, acToken, arts)
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			case errors.Is(err, services.ErrArticleNotAvailable):
				err = responseJSON(w, http.StatusForbidden, id, acToken, arts)
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			default:
				slog.Error("Can't update article", "err", err.Error())

				err = responseJSON(w, http.StatusInternalServerError, id, acToken, arts)
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		if len(arts) == 0 {
			err = responseJSONError(w, http.StatusResetContent, id, acToken, "There are no articles")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		err = responseJSON(w, http.StatusAccepted, id, acToken, arts)
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}

type deleteRequest struct {
	ArticleID int `json:"article_id"`
}

func deleteArticle(timeout time.Duration, news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(deleteRequest)

		id := r.Header.Get("uid")
		acToken := r.Header.Get("access_token")

		if id == "" {
			slog.Debug("Can't get header id is empty")

			err := responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		uid, err := strconv.Atoi(id)
		if err != nil {
			slog.Debug("Can't convert to int", "err", err.Error())

			err = responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from delete-article request", "err", err.Error())

				err = responseJSONError(w, http.StatusBadRequest, id, acToken, "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		arts, err := news.DeleteArticle(ctx, int64(uid), int64(req.ArticleID))
		if err != nil {
			switch {
			case errors.Is(err, services.ErrNoOfferedArticles):
				err = responseJSON(w, http.StatusNoContent, id, acToken, arts)
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			case errors.Is(err, services.ErrArticleNotAvailable):
				err = responseJSON(w, http.StatusAlreadyReported, id, acToken, arts)
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			default:
				slog.Error("Can't delete article", "err", err.Error())

				err = responseJSONError(w, http.StatusResetContent, id, acToken, "Internal error")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		if len(arts) == 0 {
			err = responseJSONError(w, http.StatusResetContent, id, acToken, "There are no articles")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		err = responseJSON(w, http.StatusOK, id, acToken, arts)
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}
