package handler

import (
	"path/filepath"
	"text/template"
	"time"

	"newsWebApp/app/webService/internal/forms"
	"newsWebApp/app/webService/internal/models"
)

type templateData struct {
	CurrentYear       int
	Art               *models.Art
	Articles          []models.Article
	Form              *forms.Form
	Flash             string
	AuthenticatedUser *models.User
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	abs, err := filepath.Abs(dir + "html/")
	if err != nil {
		return nil, err
	}

	pages, err := filepath.Glob(filepath.Join(abs, "*.page.html"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(abs, "*.layout.html"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(abs, "*.partial.html"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
