package handler

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
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

func render(w http.ResponseWriter,
	r *http.Request,
	service AuthService,
	name string,
	tC map[string]*template.Template,
	td *templateData,
	session *sessions.Session,
	slog *slog.Logger,
) {
	ts, ok := tC[name]
	if !ok {
		slog.Error("The template does not exists", "name", name)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	td = addDefaultData(service, td, r, session)

	err := ts.Execute(buf, td)
	if err != nil {
		slog.Error("Can't execute template", "err", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		slog.Error("Can't write to response", "err", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func addDefaultData(service AuthService, td *templateData, r *http.Request, session *sessions.Session) *templateData {
	if td == nil {
		td = &templateData{}
	}

	td.CurrentYear = time.Now().Year()
	td.Flash = session.PopString(r, "flash")
	td.AuthenticatedUser = authenticatedUser(service, r)

	return td
}

func authenticatedUser(service AuthService, r *http.Request) *models.User {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	id, userName, err := service.Parse(ctx, auth)
	if err != nil {
		return nil
	}

	user := models.User{
		ID:   id,
		Name: userName,
	}

	return &user
}
