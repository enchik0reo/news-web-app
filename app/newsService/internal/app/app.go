package app

import (
	"database/sql"
	"log/slog"
	"os"

	"newsWebApp/app/newsService/internal/config"
	"newsWebApp/app/newsService/internal/logs"
	"newsWebApp/app/newsService/internal/storage/psql"
)

type App struct {
	cfg *config.Config
	log *slog.Logger
	db  *sql.DB
}

func New() *App {
	a := App{}
	var err error

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	a.db, err = psql.New(a.cfg.Storage)
	if err != nil {
		a.log.Error("failed to create new news storage", "err", err)
		os.Exit(1)
	}

	/* sourceStor := storage.NewSourceStorage(a.db)
	articleStor := storage.NewArticleStorage(a.db) */

	return &a
}

func (a *App) MustRun() {
	a.log.Info("starting news grpc service", "env", a.cfg.Env, "port", a.cfg.GRPC.Port)

	a.mustStop()
}

func (a *App) mustStop() {
	if err := a.db.Close(); err != nil {
		a.log.Error("error closing connection to news storage", "err store", err.Error())
	}

	a.log.Info("news grpc service stoped gracefully")
}
