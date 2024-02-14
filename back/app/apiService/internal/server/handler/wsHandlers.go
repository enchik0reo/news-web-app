package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

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
