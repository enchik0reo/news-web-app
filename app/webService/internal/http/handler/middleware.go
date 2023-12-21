package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"newsWebApp/app/webService/internal/models"
	"newsWebApp/app/webService/internal/services"

	"github.com/go-chi/chi/middleware"
)

func authenticate(ctxKeyUser models.ContextKey, service AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				cookie, err := r.Cookie("refresh_token")
				if err != nil {
					if errors.Is(err, http.ErrNoCookie) {
						next.ServeHTTP(w, r)
						return
					}
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				id, userName, acsToken, refToken, err := service.Refresh(ctx, cookie.Value)
				if err != nil {
					switch {
					case errors.Is(err, services.ErrSessionNotFound):
						next.ServeHTTP(w, r)
						return
					case errors.Is(err, services.ErrInvalidValue):
						next.ServeHTTP(w, r)
						return
					default:
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
				}

				w.Header().Add("Authorization", "Bearer "+acsToken)

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

				user := models.User{
					ID:   id,
					Name: userName,
				}

				ctx = context.WithValue(r.Context(), ctxKeyUser, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			id, userName, err := service.Parse(ctx, auth)
			if err != nil {
				switch {
				case errors.Is(err, services.ErrTokenExpired):
					cookie, err := r.Cookie("refresh_token")
					if err != nil {
						if errors.Is(err, http.ErrNoCookie) {
							next.ServeHTTP(w, r)
							return
						}
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}

					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()

					id, userName, acsToken, refToken, err := service.Refresh(ctx, cookie.Value)
					if err != nil {
						switch {
						case errors.Is(err, services.ErrSessionNotFound):
							next.ServeHTTP(w, r)
							return
						case errors.Is(err, services.ErrInvalidValue):
							next.ServeHTTP(w, r)
							return
						default:
							http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
							return
						}
					}

					w.Header().Add("Authorization", "Bearer "+acsToken)

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

					user := models.User{
						ID:   id,
						Name: userName,
					}

					ctx = context.WithValue(r.Context(), ctxKeyUser, user)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				case errors.Is(err, services.ErrInvalidToken):
					next.ServeHTTP(w, r)
					return
				default:
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

			} else {
				user := models.User{
					ID:   id,
					Name: userName,
				}

				ctx = context.WithValue(r.Context(), ctxKeyUser, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		return http.HandlerFunc(fn)
	}
}

// Пока не трогаем
/* func authMw(service AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				header = r.FormValue("token")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			id, userName, err := service.Parse(ctx, header)
			if err != nil {
				switch {
				case errors.Is(err, services.ErrTokenExpired):
					cookie, err := r.Cookie("refresh_token")
					if err != nil {
						if errors.Is(err, http.ErrNoCookie) {
							http.Error(w, "authorization error, please sign-in", http.StatusBadRequest)
							return
						}
						http.Error(w, "internal error", http.StatusInternalServerError)
						return
					}

					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()

					id, userName, acsToken, refToken, err := service.Refresh(ctx, cookie.Value)
					if err != nil {
						switch {
						case errors.Is(err, services.ErrSessionNotFound):
							http.Error(w, "session expired, please sign-in", http.StatusUnauthorized)
							return
						case errors.Is(err, services.ErrInvalidValue):
							http.Error(w, "incorrect session, please sign-in", http.StatusInternalServerError)
							return
						default:
							http.Error(w, "internal error", http.StatusInternalServerError)
							return
						}
					}

					w.Header().Add("Authorization", "Bearer "+acsToken)

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

					w.Header().Add("id", fmt.Sprint(id))
					w.Header().Add("user_name", userName)

					next.ServeHTTP(w, r)

				case errors.Is(err, services.ErrInvalidToken):
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				default:
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}

			} else {
				w.Header().Add("id", fmt.Sprint(id))
				w.Header().Add("user_name", userName)

				next.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(fn)
	}
} */

func requireAuthenticatedUser(ctxKeyUser models.ContextKey, service AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			// If the user is not authenticated, redirect them to the login page and
			// return from the middleware chain so that no subsequent handlers in
			// the chain are executed.
			if authenticatedUser(ctxKeyUser, r) == nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Otherwise call the next handler in the chain.
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
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
