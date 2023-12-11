package app

import (
	"fmt"

	"newsWebApp/app/authService/internal/config"
)

type App struct {
	cfg *config.Config
	/* log         *slog.Logger
	userStor    *psql.Storage
	sessionStor *redis.Storage
	auth        *auth.Auth
	gRPCServer  *grpcsrv.Server */
}

func New() *App {
	a := App{}
	//var err error

	a.cfg = config.MustLoad()

	// TODO Инициализировать логгер

	// TODO Инициализировать users хранилище

	// TODO Инициализировать sessions хранилище

	// TODO Авторизатор (Логика авторизации)

	// TODO gRPC сервер

	return &a
}

func (a *App) Run() {

	fmt.Printf("%+v\n", a.cfg)
	// TODO Запустить сервер приложения в отдельной горутине

	// TODO Graceful shutdown
}
