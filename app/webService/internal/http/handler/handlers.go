package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"text/template"
	"time"

	"newsWebApp/app/webService/internal/forms"
	"newsWebApp/app/webService/internal/models"
	"newsWebApp/app/webService/internal/services"

	"github.com/golangcollege/sessions"
)

func home(service AuthService,
	fetcher NewsFetcher,
	tC map[string]*template.Template,
	session *sessions.Session,
	slog *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		arts, err := fetcher.FetchArticles(ctx)
		if err != nil {
			slog.Debug("Can't fetch articles", "err", err.Error())
		}

		render(w, r, service, "home.page.html", tC, &templateData{Articles: arts}, session, slog)
	}
}

func signupForm(service AuthService,
	tC map[string]*template.Template,
	session *sessions.Session,
	slog *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, service, "signup.page.html", tC, &templateData{
			Form: forms.New(nil),
		}, session, slog)
	}
}

func signup(service AuthService,
	tC map[string]*template.Template,
	session *sessions.Session,
	slog *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			slog.Error("Can't parse form", "err", err.Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		form := forms.New(r.PostForm)
		form.Required("name", "email", "password")
		form.MaxLength("name", 100)
		form.MatchesPattern("email", forms.EmailRX)
		form.MinLength("password", 6)

		if !form.Valid() {
			render(w, r, service, "signup.page.html", tC, &templateData{Form: form}, session, slog)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, err = service.SaveUser(ctx, form.Get("name"), form.Get("email"), form.Get("password"))
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserExists):
				form.Errors.Add("email", "Address is already in use")
				render(w, r, service, "signup.page.html", tC, &templateData{Form: form}, session, slog)
				return
			case errors.Is(err, services.ErrInvalidValue):
				form.Errors.Add("email", "Invalid address")
				render(w, r, service, "signup.page.html", tC, &templateData{Form: form}, session, slog)
				return
			default:
				slog.Error("Can't save user", "err", err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		session.Put(r, "flash", "Your signup was successfully. Please log in.")

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func loginForm(service AuthService,
	tC map[string]*template.Template,
	session *sessions.Session,
	slog *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, service, "login.page.html", tC, &templateData{
			Form: forms.New(nil),
		}, session, slog)
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func login(service AuthService,
	tC map[string]*template.Template,
	session *sessions.Session,
	slog *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(loginRequest)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			slog.Error("Failed to decode request", "err", err.Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		form := forms.New(r.PostForm)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, _, acsToken, refToken, err := service.LoginUser(ctx, req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserDoesntExists):
				form.Errors.Add("generic", "Email or Password is incorrect")
				render(w, r, service, "login.page.html", tC, &templateData{Form: form}, session, slog)
				return
			case errors.Is(err, services.ErrInvalidValue):
				form.Errors.Add("generic", "Email or Password is incorrect")
				render(w, r, service, "login.page.html", tC, &templateData{Form: form}, session, slog)
				return
			default:
				slog.Error("Can't login user", "err", err.Error())
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
	}
}

func logout(session *sessions.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("Authorization")
		w.Header().Del("Authorization")
	}
}

func refresh(service AuthService,
	tC map[string]*template.Template,
	session *sessions.Session,
	slog *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
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
					return
				}

				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				_, _, acsToken, refToken, err := service.Refresh(ctx, cookie.Value)
				if err != nil {
					slog.Error("Can't do service's refresh", "err", err.Error())
					return
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
			}
		}
	}
}

func suggestArticleForm(service AuthService,
	tC map[string]*template.Template,
	session *sessions.Session,
	slog *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, service, "create.page.html", tC, &templateData{
			Form: forms.New(nil),
		}, session, slog)
	}
}

type suggestRequest struct {
	Link    string `json:"link"`
	Content string `json:"content"`
}

func suggestArticle(service AuthService,
	news NewsSaver,
	tC map[string]*template.Template,
	ctxKeyArticle models.ContextKeyArticle,
	session *sessions.Session,
	slog *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(suggestRequest)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			slog.Error("Failed to decode request", "err", err.Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		form := forms.New(r.PostForm)

		if !form.Valid() {
			render(w, r, service, "create.page.html", tC, &templateData{Form: form}, session, slog)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := news.SaveArticle(ctx, 0, req.Link); err != nil {
			slog.Error("Can't save article", "err", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		///////////////////////////////////////////////////////////////////////

		art := models.Art{Link: req.Link, Content: req.Content}

		ctx = context.WithValue(r.Context(), ctxKeyArticle, art)

		http.Redirect(w, r.WithContext(ctx), "/article/suggested", http.StatusSeeOther)
	}
}

// TODO Нужно сделать получение на подобие метода login, т.е. js скриптом отправлять насервер json и тут его парсить!
func showArticle(service AuthService,
	tC map[string]*template.Template,
	ctxKeyArticle models.ContextKeyArticle,
	session *sessions.Session,
	slog *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		art, ok := r.Context().Value(ctxKeyArticle).(*models.Art)

		if !ok {
			slog.Error("Can't get article form context", "context key", ctxKeyArticle)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		render(w, r, service, "show.page.html", tC, &templateData{
			Art: art,
		}, session, slog)
	}
}
