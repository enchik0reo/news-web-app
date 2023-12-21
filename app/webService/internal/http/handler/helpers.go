package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"newsWebApp/app/webService/internal/models"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi"
	"github.com/golangcollege/sessions"
)

type cutedFileSystem struct {
	fs http.FileSystem
}

func (nfs cutedFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}

func fileServer(r chi.Router, path string, templPath string) error {
	abs, err := filepath.Abs(templPath)
	if err != nil {
		return err
	}

	if strings.ContainsAny(path, "{}*") {
		return fmt.Errorf("fileServer does not permit any URL parameters")
	}

	root := http.Dir(abs + path)

	fs := http.StripPrefix(path, http.FileServer(cutedFileSystem{root}))

	r.Get(path+"*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))

	return nil
}

func render(w http.ResponseWriter, r *http.Request, name string, tC map[string]*template.Template, td *templateData, ctxKeyUser models.ContextKey, session *sessions.Session) {
	ts, ok := tC[name]
	if !ok {
		http.Error(w, fmt.Sprintf("the template %s does not exists", name), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.Execute(buf, addDefaultData(ctxKeyUser, td, r, session))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

func addDefaultData(ctxKeyUser models.ContextKey, td *templateData, r *http.Request, session *sessions.Session) *templateData {
	if td == nil {
		td = &templateData{}
	}

	td.CurrentYear = time.Now().Year()
	td.Flash = session.PopString(r, "flash")
	td.AuthenticatedUser = authenticatedUser(ctxKeyUser, r)

	return td
}

func authenticatedUser(ctxKeyUser models.ContextKey, r *http.Request) *models.User {
	user, ok := r.Context().Value(ctxKeyUser).(*models.User)
	if !ok {
		return nil
	}
	return user
}
