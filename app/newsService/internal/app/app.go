package app

import (
	"log/slog"
	"os"

	"newsWebApp/app/newsService/internal/config"
	"newsWebApp/app/newsService/internal/logs"
	"newsWebApp/app/newsService/internal/storage/psql"
)

type App struct {
	cfg  *config.Config
	log  *slog.Logger
	stor *psql.Storage
}

func New() *App {
	a := App{}
	var err error

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	a.stor, err = psql.New(a.cfg.Storage)
	if err != nil {
		a.log.Error("failed to create new news storage", "err", err)
		os.Exit(1)
	}

	return &a
}

func (a *App) MustRun() {

}
