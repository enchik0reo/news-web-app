package handler

import (
	"context"
	"errors"
	"net/http"
	"text/template"
	"time"

	"newsWebApp/app/webService/internal/forms"
	"newsWebApp/app/webService/internal/models"
	"newsWebApp/app/webService/internal/services"

	"github.com/golangcollege/sessions"
)

type Request struct {
	UserName string `json:"user_name"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func home(fetcher NewsFetcher,
	tC map[string]*template.Template,
	ctxKeyUser models.ContextKey,
	session *sessions.Session,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		arts, err := fetcher.FetchArticles(ctx)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		render(w, r, "home.page.html", tC, &templateData{Articles: arts}, ctxKeyUser, session)
	}
}

func signupForm(tC map[string]*template.Template,
	ctxKeyUser models.ContextKey,
	session *sessions.Session,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "signup.page.html", tC, &templateData{
			Form: forms.New(nil),
		}, ctxKeyUser, session)
	}
}

func signup(service AuthService,
	tC map[string]*template.Template,
	ctxKeyUser models.ContextKey,
	session *sessions.Session,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		form := forms.New(r.PostForm)
		form.Required("name", "email", "password")
		form.MatchesPattern("email", forms.EmailRX)
		form.MinLength("password", 6)

		if !form.Valid() {
			render(w, r, "signup.page.html", tC, &templateData{Form: form}, ctxKeyUser, session)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, err = service.SaveUser(ctx, form.Get("name"), form.Get("email"), form.Get("password"))
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserExists):
				form.Errors.Add("email", "Address is already in use")
				render(w, r, "signup.page.html", tC, &templateData{Form: form}, ctxKeyUser, session)
				return
			case errors.Is(err, services.ErrInvalidValue):
				form.Errors.Add("email", "Invalid address")
				render(w, r, "signup.page.html", tC, &templateData{Form: form}, ctxKeyUser, session)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		session.Put(r, "flash", "Your signup was successfully. Please log in.")

		http.Redirect(w, r, "/login", http.StatusSeeOther)

		/* req := new(Request)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(w, "failed to decode request", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		id, err := service.SaveUser(ctx, req.UserName, req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserExists):
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case errors.Is(err, services.ErrInvalidValue):
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int64{"id": id}) */
	}
}

func loginForm(tC map[string]*template.Template,
	ctxKeyUser models.ContextKey,
	session *sessions.Session,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "login.page.html", tC, &templateData{
			Form: forms.New(nil),
		}, ctxKeyUser, session)
	}
}

func login(service AuthService,
	tC map[string]*template.Template,
	ctxKeyUser models.ContextKey,
	session *sessions.Session,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		form := forms.New(r.PostForm)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		id, userName, acsToken, refToken, err := service.LoginUser(ctx, form.Get("email"), form.Get("password"))
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserDoesntExists):
				form.Errors.Add("generic", "Email or Password is incorrect")
				render(w, r, "login.page.html", tC, &templateData{Form: form}, ctxKeyUser, session)
				return
			case errors.Is(err, services.ErrInvalidValue):
				form.Errors.Add("generic", "Email or Password is incorrect")
				render(w, r, "login.page.html", tC, &templateData{Form: form}, ctxKeyUser, session)
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

		session.Put(r, "userID", id)
		session.Put(r, "userName", userName)

		http.SetCookie(w, &ck)

		user := models.User{
			ID:   id,
			Name: userName,
		}

		ctx = context.WithValue(r.Context(), ctxKeyUser, user)

		http.Redirect(w, r.WithContext(ctx), "/article/suggest", http.StatusSeeOther)

		/* req := new(Request)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(w, "failed to decode request", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		id, userName, acsToken, refToken, err := service.LoginUser(ctx, req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserDoesntExists):
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			case errors.Is(err, services.ErrInvalidValue):
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
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

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"Access Token":  "Bearer " + acsToken,
			"Refresh Token": refToken,
		}) */
	}
}

func logout(ctxKeyUser models.ContextKey, session *sessions.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session.Remove(r, "userID")
		session.Remove(r, "userName")

		session.Put(r, "flash", "You've been logged out successfully!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func suggestArticleForm(tC map[string]*template.Template,
	ctxKeyUser models.ContextKey,
	session *sessions.Session,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "create.page.html", tC, &templateData{
			Form: forms.New(nil),
		}, ctxKeyUser, session)
	}
}

func suggestArticle(news NewsSaver,
	tC map[string]*template.Template,
	ctxKeyUser models.ContextKey,
	ctxKeyArticle models.ContextKeyArticle,
	session *sessions.Session,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)

		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		form := forms.New(r.PostForm)
		form.Required("link", "content")
		form.MaxLength("link", 200)
		form.MaxLength("content", 200)

		if !form.Valid() {
			render(w, r, "create.page.html", tC, &templateData{Form: form}, ctxKeyUser, session)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		uid := session.Get(r, "userID")

		if err := news.SaveArticle(ctx, uid.(int64), form.Get("link")); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		session.Put(r, "flash", "Article successfully suggested")

		art := models.Art{Link: form.Get("link"), Content: form.Get("content")}

		ctx = context.WithValue(r.Context(), ctxKeyArticle, art)

		http.Redirect(w, r.WithContext(ctx), "/article/suggested", http.StatusSeeOther)
	}
}

func showArticle(tC map[string]*template.Template,
	ctxKeyUser models.ContextKey,
	ctxKeyArticle models.ContextKeyArticle,
	session *sessions.Session,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		art, ok := r.Context().Value(ctxKeyArticle).(*models.Art)

		if !ok {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		render(w, r, "show.page.html", tC, &templateData{
			Art: art,
		}, ctxKeyUser, session)
	}
}

func whoami(ctxKeyUser models.ContextKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(ctxKeyUser).(*models.User)
		if !ok {
			w.Write([]byte("Please login"))
		}
		w.Write([]byte("Your name is " + user.Name))
	}
}

func ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Pong"))
	}
}
