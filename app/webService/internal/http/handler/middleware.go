package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
)

/* func authenticate(service AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				//http.Error(w, "no auth header", http.StatusNotFound)
				next.ServeHTTP(w, r)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			_, _, err := service.Parse(ctx, auth)
			if err != nil {
				switch {
				case errors.Is(err, services.ErrTokenExpired):
					cookie, err := r.Cookie("refresh_token")
					if err != nil {
						if errors.Is(err, http.ErrNoCookie) {
							//http.Error(w, "no kookie", http.StatusNotFound)
							next.ServeHTTP(w, r)
							return
						}
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}

					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()

					_, _, acsToken, refToken, err := service.Refresh(ctx, cookie.Value)
					if err != nil {
						switch {
						case errors.Is(err, services.ErrSessionNotFound):
							//http.Error(w, "session die", http.StatusNotFound)
							next.ServeHTTP(w, r)
							return
						case errors.Is(err, services.ErrInvalidToken):
							//http.Error(w, "invalid token", http.StatusNotFound)
							next.ServeHTTP(w, r)
							return
						default:
							http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
							return
						}
					}

					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]string{
						"AccessToken": acsToken,
					})

					ck := http.Cookie{
						Name:     "refresh_token",
						Domain:   r.URL.Host,
						Path:     "/",
						Value:    refToken,
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteStrictMode,
					}

					http.SetCookie(w, &ck)

					next.ServeHTTP(w, r)
					return
				case errors.Is(err, services.ErrInvalidToken):
					//http.Error(w, "invalid token", http.StatusNotFound)
					next.ServeHTTP(w, r)
					return
				default:
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

			} else {

				next.ServeHTTP(w, r)
				return
			}
		}
		return http.HandlerFunc(fn)
	}
} */

/* func requireAuthenticatedUser(service AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			if authenticatedUser(service, r) == nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
} */

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
				entry.Info("request completed",
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

func secureHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("X-Frame-Options", "deny")

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
