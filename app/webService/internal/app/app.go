package app

import "newsWebApp/app/webService/internal/config"

type App struct {
	cfg *config.Config
}

func New() *App {
	a := App{}

	a.cfg = config.MustLoad()

	return &a
}

func (a *App) Run() {

}
