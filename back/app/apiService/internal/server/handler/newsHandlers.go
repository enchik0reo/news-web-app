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

func home(timeout time.Duration, fetcher NewsFetcher, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		statusCode := 0

		defer func() {
			metrics.ObserveRequest(time.Since(start), statusCode)
		}()

		id, acToken := getInfoFromCtx(r)

		currentPage := r.URL.Query().Get("page")

		page, err := strconv.Atoi(currentPage)
		if err != nil {
			slog.Debug("Can't convert current page to int", "err", err.Error())
			statusCode = http.StatusBadRequest

			if err = responseJSONError(w, statusCode, id, acToken, "Bad request"); err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		arts, err := fetcher.FetchArticlesOnPage(ctx, int64(page))
		if err != nil {
			switch {
			case errors.Is(err, services.ErrNoPublishedArticles):
				statusCode = http.StatusNoContent

				if err = responseJSONError(w, statusCode, id, acToken, "There are no articles"); err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			case errors.Is(err, context.DeadlineExceeded):
				slog.Debug("Can't fetch articles", "err", err.Error())
				statusCode = http.StatusInternalServerError

				if err = responseJSONError(w, statusCode, id, acToken, "Internal server error"); err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			default:
				slog.Error("Can't fetch articles", "err", err.Error())
				statusCode = http.StatusInternalServerError

				if err = responseJSONError(w, statusCode, id, acToken, "Internal server error"); err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		statusCode = http.StatusOK

		if err = responseJSON(w, statusCode, id, acToken, arts); err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}

func userArticles(timeout time.Duration, news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, acToken := getInfoFromCtx(r)

		if id == 0 {
			slog.Debug("Can't get user id")

			err := responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		arts, err := news.GetArticlesByUid(ctx, id)
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
		defer r.Body.Close()
		req := new(addRequest)

		id, acToken := getInfoFromCtx(r)

		if id == 0 {
			slog.Debug("Can't get user id")

			err := responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
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

		arts, err := news.SaveArticle(ctx, id, req.Link)
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
			case errors.Is(err, services.ErrInvalidUrl):
				err = responseJSON(w, http.StatusMethodNotAllowed, id, acToken, arts)
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
	ArticleID int64  `json:"article_id"`
	Link      string `json:"link"`
}

func updateArticle(timeout time.Duration, news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := new(updateRequest)

		id, acToken := getInfoFromCtx(r)

		if id == 0 {
			slog.Debug("Can't get user id")

			err := responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
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

		arts, err := news.UpdateArticle(ctx, id, req.ArticleID, req.Link)
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
			case errors.Is(err, services.ErrInvalidUrl):
				err = responseJSON(w, http.StatusMethodNotAllowed, id, acToken, arts)
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
	ArticleID int64 `json:"article_id"`
}

func deleteArticle(timeout time.Duration, news UserNewsService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := new(deleteRequest)

		id, acToken := getInfoFromCtx(r)

		if id == 0 {
			slog.Debug("Can't get user id")

			err := responseJSONError(w, http.StatusInternalServerError, id, acToken, "Internal error")
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

		arts, err := news.DeleteArticle(ctx, id, req.ArticleID)
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