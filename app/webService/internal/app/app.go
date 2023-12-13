package app

import (
	"log/slog"

	"newsWebApp/app/webService/internal/config"
	"newsWebApp/app/webService/internal/logs"
)

type App struct {
	cfg *config.Config
	log *slog.Logger
}

func New() *App {
	a := App{}

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	return &a
}
