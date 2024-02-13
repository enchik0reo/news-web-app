package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"newsWebApp/app/apiService/internal/services"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

func authenticate(refTokTTL time.Duration, service AuthService, slog *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				slog.Debug("Can't authenticate user, access token is empty")

				err := responseJSONError(w, http.StatusNotFound, "", "", "Empty Authorization")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			_, _, err := service.Parse(ctx, auth)
			if err != nil {
				slog.Debug("Can't authenticate user", "error", err.Error())

				err = responseJSONError(w, http.StatusUnauthorized, "", "", "Authorization expired")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			} else {
				next.ServeHTTP(w, r)
				return
			}
		}
		return http.HandlerFunc(fn)
	}
}

func refresh(refTokTTL time.Duration, service AuthService, slog *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				slog.Debug("No auth header")
				next.ServeHTTP(w, r)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			id, _, err := service.Parse(ctx, auth)
			if err != nil {
				switch {
				case errors.Is(err, services.ErrTokenExpired):
					cookie, err := r.Cookie("refresh_token")
					if err != nil {
						if errors.Is(err, http.ErrNoCookie) {
							slog.Debug("No cookie")
							next.ServeHTTP(w, r)
							return
						} else {
							slog.Error("Can't get cookie", "err", err.Error())
							next.ServeHTTP(w, r)
							return
						}
					}

					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()

					id, _, acsToken, refToken, err := service.Refresh(ctx, cookie.Value)
					if err != nil {
						slog.Debug("Can't do refresh tokens", "err", err.Error())
						next.ServeHTTP(w, r)
						return
					}

					r.Header.Set("Authorization", "Bearer "+acsToken)
					r.Header.Set("uid", fmt.Sprint(id))
					r.Header.Set("access_token", acsToken)

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

					next.ServeHTTP(w, r)
					return
				default:
					slog.Debug("Can't parse access token", "err", err.Error())
					next.ServeHTTP(w, r)
					return
				}
			} else {
				r.Header.Set("uid", fmt.Sprint(id))

				next.ServeHTTP(w, r)
				return
			}
		}
		return http.HandlerFunc(fn)
	}
}

func corsSettings() func(next http.Handler) http.Handler {
	h := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3003"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowedHeaders:   []string{"Content-Type", "Set-Cookie", "Authorization", "id"},
		ExposedHeaders:   []string{"Content-Type", "Set-Cookie", "Authorization", "id"},
		AllowCredentials: true,
	})

	return h
}

func loggerMw(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := log.With(
			slog.String("component", "middleware/logger"),
		)

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				entry.Debug("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
